package atom

import (
	"sync/atomic"
	"unsafe"

	"jsouthworth.net/go/dyn"
)

type Atom struct {
	state *interface{}
}

func New(s interface{}) *Atom {
	return &Atom{state: &s}
}

func (a *Atom) Deref() interface{} {
	out := a.load()
	return *out
}

func (a *Atom) Swap(fn interface{}, args ...interface{}) interface{} {
	for {
		old := a.load()
		new := dyn.Apply(fn, dyn.PrependArg(*old, args...)...)
		if a.compareAndSwap(old, &new) {
			return new
		}
	}
}

func (a *Atom) Reset(new interface{}) interface{} {
	atomic.StorePointer(a.loc(), unsafe.Pointer(&new))
	return new
}

func (a *Atom) loc() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(&a.state))
}

func (a *Atom) load() *interface{} {
	outPtr := atomic.LoadPointer(a.loc())
	return (*interface{})(outPtr)
}

func (a *Atom) compareAndSwap(old, new *interface{}) bool {
	return atomic.CompareAndSwapPointer(a.loc(),
		unsafe.Pointer(old), unsafe.Pointer(new))
}
