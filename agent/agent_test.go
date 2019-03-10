package agent

import (
	"fmt"
	"testing"
	"time"

	"jsouthworth.net/go/immutable/hashmap"
	"jsouthworth.net/go/seq"
)

func TestOne(t *testing.T) {
	agt := New(1)
	a := agt.Deref()
	if a != 1 {
		t.Fatalf("got %v, wanted %v\n", a, 1)
	}
	agt.Send(func(cur int) int {
		return 2
	})
	for {
		b := agt.Deref()
		if b == 2 {
			break
		}
	}
}

func TestMany(t *testing.T) {
	agt := New(0)
	expected := 0
	for i := 0; i < 100; i++ {
		expected += i
		go func(i int) {
			agt.Send(func(cur int, next int) int {
				return cur + next
			}, i)
		}(i)
	}
	done := make(chan struct{})
	go func() {
		for {
			got := agt.Deref()
			if got == expected {
				break
			}
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout while waiting for agent to finish")
	}
}

func TestRing(t *testing.T) {
	/* Replicate the example from the clojure docs
	;; This example is an implementation of the
	;; send-a-message-around-a-ring test. A chain of n agents
	;; is created, then a sequence of m actions are
	;; dispatched to the head of the chain and relayed
	;; through it.
		   (defn relay [x i]
		     (when (:next x)
		       (send (:next x) relay i))
		     (when (and (zero? i) (:report-queue x))
		       (.put (:report-queue x) i))
		     x)

		   (defn run [m n]
		     (let [q (new java.util.concurrent.SynchronousQueue)
		           hd (reduce (fn [next _] (agent {:next next}))
		                      (agent {:report-queue q}) (range (dec m)))]
		       (doseq [i (reverse (range n))]
		         (send hd relay i))
		       (.take q)))

		   ; 1 million message sends:
		   (time (run 1000 1000))
	*/
	var relay func(x *hashmap.Map, i int) interface{}
	relay = func(x *hashmap.Map, i int) interface{} {
		switch {
		case x.Contains(":next"):
			x.At(":next").(*Agent).Send(relay, i)
		case i == 0 && x.Contains(":report-queue"):
			x.At(":report-queue").(chan int) <- i
		}
		return x
	}

	run := func(m, n int) {
		q := make(chan int)
		hd := seq.Reduce(
			func(next *Agent, _ int) *Agent {
				return New(hashmap.New(":next", next))
			},
			New(hashmap.New(":report-queue", q)),
			seq.RangeUntil(m-1)).(*Agent)
		for i := n; i >= 0; i-- {
			hd.Send(relay, i)
		}
		<-q

	}
	start := time.Now()
	run(1000, 1000)
	t.Log("1 million message sends took:", time.Since(start))

}

func ExampleAgent_Send_ring() {
	// This example is an implementation of the
	// send-a-message-around-a-ring test. A chain of n agents
	// is created, then a sequence of m actions are
	// dispatched to the head of the chain and relayed
	// through it.
	var relay func(x *hashmap.Map, i int) interface{}
	relay = func(x *hashmap.Map, i int) interface{} {
		switch {
		case x.Contains(":next"):
			x.At(":next").(*Agent).Send(relay, i)
		case i == 0 && x.Contains(":report-queue"):
			x.At(":report-queue").(chan int) <- i
		}
		return x
	}

	run := func(m, n int) {
		q := make(chan int)
		hd := seq.Reduce(
			func(next *Agent, _ int) *Agent {
				return New(hashmap.New(":next", next))
			},
			New(hashmap.New(":report-queue", q)),
			seq.RangeUntil(m-1)).(*Agent)
		for i := n; i >= 0; i-- {
			hd.Send(relay, i)
		}
		fmt.Println(<-q)

	}
	run(1000, 1000)
	//Output: 0
}
