package agent

import (
	"github.com/jsouthworth/etm/atom"
)

type agentRequest struct {
	fn   atom.XfrmFn
	args []interface{}
}

type Agent struct {
	state *atom.Atom
	ch    chan agentRequest
}

func New(s interface{}) *Agent {
	a := &Agent{state: atom.New(s), ch: make(chan agentRequest, 100)}
	go a.process()
	return a
}

func (a *Agent) Send(fn atom.XfrmFn, args ...interface{}) *Agent {
	a.ch <- agentRequest{fn: fn, args: args}
	return a
}

func (a *Agent) process() {
	for {
		r := <-a.ch
		a.state.Swap(r.fn, r.args...)
	}
}
