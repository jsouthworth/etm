package ref

import (
	"sync/atomic"
	"unsafe"
)

type Ref struct {
	data unsafe.Pointer
}

func Make(data unsafe.Pointer) Ref {
	return Ref{data: data}
}

func (r *Ref) Load() unsafe.Pointer {
	return atomic.LoadPointer(&r.data)
}

func (r *Ref) CompareAndSwap(old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&r.data, old, new)
}

func (r *Ref) Set(new unsafe.Pointer) {
	atomic.StorePointer(&r.data, new)
}
