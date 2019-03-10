// Package agent implements a mechanism to manage a single piece of
// state. It operates asynchronously to the rest of the program. It does
// not use a blocking goroutine for this, instead it spins up a new
// goroutine as needed and terminates the routine when its queue of work
// is empty. Accessing the value of an agent does not require
// coordination with other accessors or with updaters.
package agent

import (
	"unsafe"

	"jsouthworth.net/go/etm/atom"
	"jsouthworth.net/go/etm/unsafe/ref"
	"jsouthworth.net/go/immutable/queue"
)

// Agent is a mechanism to manage a single piece of state.
type Agent struct {
	state *atom.Atom
	queue ref.Ref
}

// New returns a new agent with an initial value of s.
func New(s interface{}) *Agent {
	return &Agent{
		state: atom.New(s),
		queue: ref.Make(unsafe.Pointer(queue.Empty())),
	}
}

// Send dispatches an action. It returns immediately and the value
// managed by the agent will be updated asynchronusly. Send takes a
// function of the type func(old aT, args...) rT where aT is the old
// type of the agent and rT is the desired type of the atom.
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

// Deref returns the current value of the agent.
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
