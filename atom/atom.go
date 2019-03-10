package atom

import (
	"unsafe"

	"jsouthworth.net/go/dyn"
	"jsouthworth.net/go/etm/unsafe/ref"
)

type Atom struct {
	state ref.Ref
}

func New(s interface{}) *Atom {
	return &Atom{state: ref.Make(unsafe.Pointer(&s))}
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
