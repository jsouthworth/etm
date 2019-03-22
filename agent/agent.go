// Package agent implements a mechanism to manage a single piece of
// state. It operates asynchronously to the rest of the program. It does
// not use a blocking goroutine for this, instead it spins up a new
// goroutine as needed and terminates the routine when its queue of work
// is empty. Accessing the value of an agent does not require
// coordination with other accessors or with updaters.
package agent

import (
	"jsouthworth.net/go/dyn"
	"jsouthworth.net/go/etm/atom"
	"jsouthworth.net/go/etm/internal/jobq"
)

// Agent is a mechanism to manage a single piece of state.
type Agent struct {
	state *atom.Atom
	queue *jobq.Queue
}

// New returns a new agent with an initial value of s.
func New(s interface{}) *Agent {
	return &Agent{
		state: atom.New(s),
		queue: jobq.New(runAction),
	}
}

// Send dispatches an action. It returns immediately and the value
// managed by the agent will be updated asynchronusly. Send takes a
// function of the type func(old aT, args...) rT where aT is the old
// type of the agent and rT is the desired type of the atom.
func (a *Agent) Send(fn interface{}, args ...interface{}) *Agent {
	action := &agentRequest{state: a.state, fn: fn, args: args}
	a.queue.Enqueue(action)
	return a
}

// Deref returns the current value of the agent.
func (a *Agent) Deref() interface{} {
	return a.state.Deref()
}

func runAction(val interface{}) {
	action := val.(*agentRequest)
	action.state.Swap(action.fn, action.args...)
}

type agentRequest struct {
	state *atom.Atom
	fn    interface{}
	args  []interface{}
}

// Watch adds function to be called when the value of the atom changes.
//
// Watchers must be functions of the following form:
// func(key kT, atom *Atom, old oT, new nT)
// Watchers may return a value or not but any returned value is ignored.
// The key type must be the type of the key passed in when the watcher
// is added, the Value types must be the type of the atom. If the atom
// can take arbitrary types then the watcher should take type interface{}.
//
// Watchers are only called when the value actually changes not on every
// Swap.
//
// All watchers are called asynchronously when the atom changes, this means
// that one should not deref the atom in the watcher but should used the
// passed in old and new values. Dispatches to the watchers are queued so
// the set of watchers for a given update are called then the set for the
// next and so on.
func (a *Agent) Watch(key interface{}, fn interface{}, args ...interface{}) *Agent {
	a.state.Watch(key, &agentWatcher{fn: fn, agent: a}, args...)
	return a
}

// Ignore removes the watcher with the passed in key so that on the
// next update it will not be in the watcher set. This takes effect
// immediately so if a watcher removes its self, the next update to the
// atom will not contain the watcher.
func (a *Agent) Ignore(key interface{}) *Agent {
	a.state.Ignore(key)
	return a
}

type agentWatcher struct {
	fn    interface{}
	agent *Agent
}

func (w *agentWatcher) Apply(args ...interface{}) interface{} {
	args[1] = w.agent // replace the atom with this agent
	return dyn.Apply(w.fn, args...)
}
