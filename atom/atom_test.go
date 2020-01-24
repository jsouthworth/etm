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
		t.Log("there was no contention on the atom",
			total, runtime.GOMAXPROCS(-1))
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

func TestWatch(t *testing.T) {
	const n = 10000
	var wg sync.WaitGroup
	watcher := func(key string, a *Atom, old int, new int) {
		wg.Done()
	}
	wg.Add(5 * n)
	a := New(0).Watch("foo", watcher).
		Watch("bar", watcher).
		Watch("baz", watcher).
		Watch("quux", watcher).
		Watch("quuz", watcher)
	for i := 0; i <= n; i++ {
		a.Reset(i)
	}
	wg.Wait()
}

func TestWatchUpdate(t *testing.T) {
	const n = 10000
	const numWatcher = 5

	bval := 0
	for i := 0; i <= n; i++ {
		bval = bval + i
	}
	bval = bval * 5

	done := make(chan struct{})
	var wg sync.WaitGroup

	b := New(0)

	watcher := func(key string, a *Atom, old int, new int) {
		b.Swap(func(old int) int { return old + new })
		wg.Done()
	}
	bwatcher := func(key int, a *Atom, old int, new int) {
		if new == bval {
			close(done)
		}
	}
	b.Watch(1, bwatcher)

	wg.Add(n * numWatcher)

	a := New(0).Watch("foo", watcher).
		Watch("bar", watcher).
		Watch("baz", watcher).
		Watch("quux", watcher).
		Watch("quuz", watcher)

	for i := 0; i <= n; i++ {
		a.Reset(i)
	}

	wg.Wait()
	<-done
}

func TestIgnore(t *testing.T) {
	const n = 1000
	a := New(0)
	count := New(0)
	done := make(chan struct{})
	watcher := func(key string, a *Atom, old int, new int) {
		count.Swap(func(old int) int { return old + new })
		if new == 10 {
			a.Ignore(key)
			close(done)
		}
	}
	a.Watch("foo", watcher)
	for i := 0; i <= n; i++ {
		a.Reset(i)
	}
	<-done
	cnt := count.Deref().(int)
	if cnt != 55 {
		t.Fatal("ignore failed got:", cnt)
	}
}

func TestWatchArgs(t *testing.T) {
	const n = 10000
	ch := make(chan string, 1)
	var wg sync.WaitGroup
	watcher := func(key string, a *Atom, old int, new int, in string) {
		ch <- in
		wg.Done()
	}
	wg.Add(5 * n)
	a := New(0).Watch("foo", watcher, "a").
		Watch("bar", watcher, "a").
		Watch("baz", watcher, "a").
		Watch("quux", watcher, "a").
		Watch("quuz", watcher, "a")
	for i := 0; i <= n; i++ {
		a.Reset(i)
	}
	for i := 0; i < 5*n; i++ {
		<-ch
	}
	wg.Wait()
}
