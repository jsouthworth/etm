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
	"jsouthworth.net/go/etm/internal/unsafe/ref"
)

// Atom is a mechanism to manage a single piece of shared state synchronously.
type Atom struct {
	state ref.Ref
}

// New returns a new atom with an initial value of s.
func New(s interface{}) *Atom {
	return &Atom{state: ref.Make(unsafe.Pointer(&s))}
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
func (a *Atom) Swap(fn interface{}, args ...interface{}) interface{} {
	for {
		old := a.load()
		new := dyn.Apply(fn, dyn.PrependArg(*old, args...)...)
		if a.compareAndSwap(old, &new) {
			return new
		}
	}
}

// Reset forcibly updates the value of the atom to new.
func (a *Atom) Reset(new interface{}) interface{} {
	a.state.Set(unsafe.Pointer(&new))
	return new
}

func (a *Atom) load() *interface{} {
	return (*interface{})(a.state.Load())
}

func (a *Atom) compareAndSwap(old, new *interface{}) bool {
	return a.state.CompareAndSwap(
		unsafe.Pointer(old), unsafe.Pointer(new))
}
