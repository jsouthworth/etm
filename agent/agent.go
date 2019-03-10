package agent

import (
	"unsafe"

	"jsouthworth.net/go/etm/atom"
	"jsouthworth.net/go/etm/unsafe/ref"
	"jsouthworth.net/go/immutable/queue"
)

type Agent struct {
	state *atom.Atom
	queue ref.Ref
}

func New(s interface{}) *Agent {
	return &Agent{
		state: atom.New(s),
		queue: ref.Make(unsafe.Pointer(queue.Empty())),
	}
}

func (a *Agent) Send(fn interface{}, args ...interface{}) *Agent {
	action := &agentRequest{fn: fn, args: args}
	var prev unsafe.Pointer
	for {
		prev = a.queue.Load()
		new := unsafe.Pointer((*queue.Queue)(prev).Push(action))
		if a.queue.CompareAndSwap(prev, new) {
			break
		}
	}
	if (*queue.Queue)(prev).Length() == 0 {
		go a.run()
	}
	return a
}

func (a *Agent) Deref() interface{} {
	return a.state.Deref()
}

func (a *Agent) run() {
	for {
		action := (*queue.Queue)(a.queue.Load()).First().(*agentRequest)
		a.state.Swap(action.fn, action.args...)
		var next *queue.Queue
		for {
			prev := a.queue.Load()
			next = (*queue.Queue)(prev).Pop()
			if a.queue.CompareAndSwap(prev, unsafe.Pointer(next)) {
				break
			}
		}
		if next.Length() == 0 {
			break
		}
	}
}

type agentRequest struct {
	fn   interface{}
	args []interface{}
}
