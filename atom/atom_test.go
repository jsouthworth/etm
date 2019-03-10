package atom

import (
	"runtime"
	"sync"
	"testing"
)

func TestSwap(t *testing.T) {
	a := New(1)
	b := a.Deref()
	if b != 1 {
		t.Fatalf("got %v, wanted %v\n", b, 1)
	}
	a.Swap(func(cur int) int {
		return 2
	})
	b = a.Deref()
	if b != 2 {
		t.Fatalf("got %v, wanted %v\n", b, 2)
	}
}

func TestConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	wait := make(chan struct{})
	executed := make(chan int, 100)
	a := New(1)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			<-wait
			count := 0
			a.Swap(func(cur int) int {
				count++
				return i + cur
			})
			executed <- count
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(executed)
	}()

	close(wait)

	total := 0
	for count := range executed {
		total += count
	}
	if total <= 100 && runtime.GOMAXPROCS(-1) > 1 {
		t.Fatal("there was no contention on the atom")
	}

	val := a.Deref()
	if val != 4951 {
		t.Fatalf("didn't get expected result, got %s, wanted 4951", val)
	}
}

func TestReset(t *testing.T) {
	a := New(1)
	b := a.Deref()
	if b != 1 {
		t.Fatalf("got %v, wanted %v\n", b, 1)
	}
	a.Reset(2)
	b = a.Deref()
	if b != 2 {
		t.Fatalf("got %v, wanted %v\n", b, 2)
	}
}
