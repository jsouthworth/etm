// Package mpscq implements a Multi-Producer, Single-Consumer
// Lock-free Queue.  This algorithm is based on Dmitry Vyukov's
// Non-intrusive MPSC node-based queue.
// http://www.1024cores.net/home/lock-free-algorithms/queues/non-intrusive-mpsc-node-based-queue
package mpscq

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	next *node
	val  interface{}
}

type Queue struct {
	head *node
	tail *node
}

func New() *Queue {
	stub := node{}
	return &Queue{
		head: &stub,
		tail: &stub,
	}
}

func (q *Queue) Push(val interface{}) {
	n := &node{
		val: val,
	}
	prev := (*node)(atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&q.head)),
		unsafe.Pointer(n),
	))
	atomic.StorePointer(
		(*unsafe.Pointer)(unsafe.Pointer(&prev.next)),
		unsafe.Pointer(n),
	)
}

func (q *Queue) Pop() (interface{}, bool) {
	tail := q.tail
	next := (*node)(atomic.LoadPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&tail.next)),
	))
	if next != nil {
		out := next.val
		next.val = nil
		atomic.StorePointer(
			(*unsafe.Pointer)(unsafe.Pointer(&q.tail)),
			unsafe.Pointer(next),
		)
		return out, false
	}
	return nil, true
}
