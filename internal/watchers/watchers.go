package watchers

import (
	"sync"
	"unsafe"

	"jsouthworth.net/go/dyn"
	"jsouthworth.net/go/etm/internal/jobq"
	"jsouthworth.net/go/etm/internal/unsafe/ref"
	"jsouthworth.net/go/immutable/hashmap"
)

type Watchers struct {
	watchers ref.Ref
	queue    *jobq.Queue
}

func New() *Watchers {
	return &Watchers{
		watchers: ref.Make(unsafe.Pointer(hashmap.Empty())),
		queue:    jobq.New(notifyWatchers),
	}
}

func (w *Watchers) Notify(ref, old, new interface{}) {
	watchers := w.getWatchers()
	if watchers.Length() == 0 {
		return
	}
	if dyn.Equal(old, new) {
		return
	}
	w.queue.Enqueue(&watcherJob{
		old:      old,
		new:      new,
		watchers: w,
		ref:      ref,
	})
}

func (w *Watchers) Add(key interface{}, watch *Watcher) {
	for {
		old := w.getWatchers()
		new := old.Assoc(key, watch)
		if w.watchers.CompareAndSwap(
			unsafe.Pointer(old),
			unsafe.Pointer(new)) {
			return
		}
	}
}

func (w *Watchers) Delete(key interface{}) {
	for {
		old := w.getWatchers()
		new := old.Delete(key)
		if w.watchers.CompareAndSwap(
			unsafe.Pointer(old),
			unsafe.Pointer(new)) {
			return
		}
	}
}

func (w *Watchers) getWatchers() *hashmap.Map {
	return (*hashmap.Map)(w.watchers.Load())
}

type Watcher struct {
	Fn   interface{}
	Args []interface{}
}

func (w *Watcher) Apply(args ...interface{}) interface{} {
	var fnargs []interface{}
	if len(w.Args) == 0 {
		fnargs = args
	} else {
		fnargs = make([]interface{}, len(w.Args)+len(args))
		copy(fnargs, args)
		copy(fnargs[len(args):], w.Args)
	}
	return dyn.Apply(w.Fn, fnargs...)
}

type watcherJob struct {
	old, new interface{}
	watchers *Watchers
	ref      interface{}
}

func notifyWatchers(val interface{}) {
	job := val.(*watcherJob)
	watchers := job.watchers.getWatchers()
	switch watchers.Length() {
	case 0:
	case 1:
		// If there is only one watcher don't incur
		// the overhead of spinning up a goroutine
		watchers.Range(func(key interface{}, w *Watcher) {
			dyn.Apply(w, key, job.ref, job.old, job.new)
		})
	default:
		var wg sync.WaitGroup
		watchers.Range(func(key interface{}, w *Watcher) {
			wg.Add(1)
			go func() {
				dyn.Apply(w, key, job.ref, job.old, job.new)
				wg.Done()
			}()
		})
		wg.Wait()
	}
}
