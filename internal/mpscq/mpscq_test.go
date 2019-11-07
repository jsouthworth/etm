package mpscq

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestQueue(t *testing.T) {
	q := New()
	q.Push("foo")
	q.Push("bar")
	val, empty := q.Pop()
	if val != "foo" || empty {
		t.Fatal("didn't get expected value foo, got:", val)
	}
	val, empty = q.Pop()
	if val != "bar" || empty {
		t.Fatal("didn't get expected value bar, got:", val)
	}
	val, empty = q.Pop()
	if !empty {
		t.Fatal("didn't get empty queue, got:", val)
	}
}

type atomicBool struct {
	state int32
}

func (b *atomicBool) boolToInt(val bool) int32 {
	if val {
		return 1
	}
	return 0
}

func (b *atomicBool) intToBool(val int32) bool {
	return val != 0
}

func newAtomicBool(init bool) *atomicBool {
	b := &atomicBool{}
	b.state = b.boolToInt(init)
	return b
}

func (b *atomicBool) Set(new bool) {
	atomic.StoreInt32(&b.state, b.boolToInt(new))
}

func (b *atomicBool) Get() bool {
	return b.intToBool(atomic.LoadInt32(&b.state))
}

func (b *atomicBool) CAS(old, new bool) bool {
	return atomic.CompareAndSwapInt32(
		&b.state,
		b.boolToInt(old),
		b.boolToInt(new),
	)
}

func TestQueueLong(t *testing.T) {
	num := 1000000
	q := New()

	start := make(chan struct{})
	done := make(chan int)
	var wg sync.WaitGroup
	wg.Add(num)

	running := newAtomicBool(false)
	for i := 0; i < num; i++ {
		go func(i int) {
			<-start
			q.Push(i)
			isRunning := running.Get()
			if !isRunning && running.CAS(isRunning, true) {
				wg.Add(1)
				go func() {
					var out int
					val, empty := q.Pop()
					for !empty {
						out += val.(int)
						val, empty = q.Pop()
					}
					running.Set(false)
					done <- out
					wg.Done()
				}()
			}
			wg.Done()
		}(i)
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	close(start)
	var acc int
	for val := range done {
		acc += val
	}

	var expected int
	for i := 0; i < num; i++ {
		expected += i
	}

	if acc != expected {
		t.Fatalf("didn't get expected value, got %v, wanted %v\n",
			acc, expected)
	}
}
