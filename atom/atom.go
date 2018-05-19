package atom

import (
	"sync/atomic"
	"unsafe"
)

type XfrmFn func(interface{}, ...interface{}) interface{}

type Atom struct {
	state *interface{}
}

func New(s interface{}) *Atom {
	return &Atom{state: &s}
}

func (a *Atom) Deref() interface{} {
	return *a.state
}

func (a *Atom) compareAndSwap(old, new *interface{}) bool {
	loc := (*unsafe.Pointer)(unsafe.Pointer(&a.state))
	return atomic.CompareAndSwapPointer(loc, unsafe.Pointer(old), unsafe.Pointer(new))
}

func (a *Atom) Swap(fn XfrmFn, args ...interface{}) interface{} {
	for {
		old := a.state
		new := fn(*old, args...)
		if a.compareAndSwap(old, &new) {
			return new
		}
	}
}

func (a *Atom) Reset(new interface{}) interface{} {
	loc := (*unsafe.Pointer)(unsafe.Pointer(&a.state))
	atomic.StorePointer(loc, unsafe.Pointer(&new))
	return new
}
