package jobq

import (
	"sync/atomic"

	"jsouthworth.net/go/etm/internal/mpscq"
)

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

type Queue struct {
	running   *atomicBool
	q         *mpscq.Queue
	processFn func(interface{})
}

func New(process func(interface{})) *Queue {
	return &Queue{
		running:   newAtomicBool(false),
		q:         mpscq.New(),
		processFn: process,
	}
}

func (q *Queue) Enqueue(value interface{}) *Queue {
	q.q.Push(value)
	isRunning := q.running.Get()
	if !isRunning && q.running.CAS(isRunning, true) {
		go q.process()
	}
	return q
}

func (q *Queue) process() {
	val, empty := q.q.Pop()
	for !empty {
		q.processFn(val)
		val, empty = q.q.Pop()
	}
	q.running.Set(false)
}
