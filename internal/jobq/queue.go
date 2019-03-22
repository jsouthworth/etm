package jobq

import (
	"unsafe"

	"jsouthworth.net/go/etm/internal/unsafe/ref"
	"jsouthworth.net/go/immutable/queue"
)

type Queue struct {
	ref       ref.Ref
	processFn func(interface{})
}

func New(process func(interface{})) *Queue {
	return &Queue{
		ref:       ref.Make(unsafe.Pointer(queue.Empty())),
		processFn: process,
	}
}

func (q *Queue) Enqueue(value interface{}) *Queue {
	var prev unsafe.Pointer
	for {
		prev = q.ref.Load()
		new := unsafe.Pointer((*queue.Queue)(prev).Push(value))
		if q.ref.CompareAndSwap(prev, new) {
			break
		}
	}
	if (*queue.Queue)(prev).Length() == 0 {
		go q.process()
	}
	return q
}

func (q *Queue) process() {
	for {
		val := (*queue.Queue)(q.ref.Load()).First()
		q.processFn(val)
		var next *queue.Queue
		for {
			prev := q.ref.Load()
			next = (*queue.Queue)(prev).Pop()
			if q.ref.CompareAndSwap(prev, unsafe.Pointer(next)) {
				break
			}
		}
		if next.Length() == 0 {
			break
		}
	}
}
