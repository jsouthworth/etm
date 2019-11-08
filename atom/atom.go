// Package atom implements a mechanism to manage a single piece of
// shared state synchronously. Updates are coordinated and serialized
// between updaters. Updates are done using a compare and swap
// loop. Access to the state happens in an uncoordinated manner meaning
// it is non-blocking to access the state of the atom. All access and
// updates are race free.
package atom

import (
	"unsafe"

	"jsouthworth.net/go/dyn"
	"jsouthworth.net/go/etm/internal/genfn"
	"jsouthworth.net/go/etm/internal/unsafe/ref"
	"jsouthworth.net/go/etm/internal/watchers"
)

// Atom is a mechanism to manage a single piece of shared state synchronously.
type Atom struct {
	state    ref.Ref
	watchers *watchers.Watchers
}

// New returns a new atom with an initial value of s.
func New(s interface{}) *Atom {
	return &Atom{
		state:    ref.Make(unsafe.Pointer(&s)),
		watchers: watchers.New(),
	}
}

// Deref returns the current value of the atom.
func (a *Atom) Deref() interface{} {
	out := a.load()
	return *out
}

// Swap updates the atom synchronously. The atom's value will be updated
// when swap returns. Swap takes a function that is of the form
// func(old aT, args...) rT where aT is the old type of the atom and rT
// is the desired type of the atom.
//
// Passing a func(...interface{})interface{} avoids reflect based
// function application allow faster execution at the expense of
// some clarity.
func (a *Atom) Swap(fn interface{}, args ...interface{}) interface{} {
	args = dyn.PrependArg(nil, args...)
	f := genfn.MakeGeneric(fn)
	for {
		old := a.load()
		args[0] = *old
		new := f(args...)
		if a.compareAndSwap(old, &new) {
			a.notifyWatchers(*old, new)
			return new
		}
	}
}

// Reset forcibly updates the value of the atom to new.
func (a *Atom) Reset(new interface{}) interface{} {
	old := a.load()
	a.state.Set(unsafe.Pointer(&new))
	a.notifyWatchers(*old, new)
	return new
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
//
// Passing a func(...interface{})interface{} avoids reflect based
// function application allow faster execution at the expense of
// some clarity.
func (a *Atom) Watch(key interface{}, fn interface{}, args ...interface{}) *Atom {
	f := genfn.MakeGeneric(fn)
	watcher := &watchers.Watcher{
		Fn:   f,
		Args: args,
	}
	a.watchers.Add(key, watcher)
	return a
}

// Ignore removes the watcher with the passed in key so that on the
// next update it will not be in the watcher set. This takes effect
// immediately so if a watcher removes its self, the next update to the
// atom will not contain the watcher.
func (a *Atom) Ignore(key interface{}) *Atom {
	a.watchers.Delete(key)
	return a
}

func (a *Atom) load() *interface{} {
	return (*interface{})(a.state.Load())
}

func (a *Atom) compareAndSwap(old, new *interface{}) bool {
	return a.state.CompareAndSwap(
		unsafe.Pointer(old), unsafe.Pointer(new))
}

func (a *Atom) notifyWatchers(old, new interface{}) {
	a.watchers.Notify(a, old, new)
}
