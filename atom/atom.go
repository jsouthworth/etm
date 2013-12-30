package atom

import (
	"unsafe"
	"sync/atomic"
)

type XfrmFn func(interface{}, ...interface{})interface{}

type Atom struct {
	state unsafe.Pointer
}

func New(s interface{}) *Atom {
	return &Atom{state: unsafe.Pointer(&s)}
}

func (a *Atom) Deref() interface{} {
	return *(*interface{})(a.state)
}

func (a *Atom) Get() *interface{} {
	return (*interface{})(a.state)
}

func (a *Atom) CompareAndSwap(old, new *interface{}) bool {
	return atomic.CompareAndSwapPointer(&a.state, unsafe.Pointer(old), unsafe.Pointer(new))
}

func (a *Atom) Swap(fn XfrmFn, args ...interface{}) interface{} {
	for {
		old := (*interface{})(a.state)
		new := fn(*old, args...)
		if a.CompareAndSwap(old, &new) {
			return new
		}
	}
}