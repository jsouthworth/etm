// Package ref implements primatives used by the other abstractions in
// this library. It is a simple wrapper around go's atomic unsafe.Pointer
// operations that make them easier to use.
package ref

import (
	"sync/atomic"
	"unsafe"
)

// Ref is an atomic reference type backed by an unsafe.Pointer
type Ref struct {
	data unsafe.Pointer
}

// Make creates a new ref with the initial data as the passed in pointer.
func Make(data unsafe.Pointer) Ref {
	return Ref{data: data}
}

// Load returns the atomic load of the current pointer.
func (r *Ref) Load() unsafe.Pointer {
	return atomic.LoadPointer(&r.data)
}

// CompareAndSwap swaps the value for the new value if the current value
// is equal to old.
func (r *Ref) CompareAndSwap(old, new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&r.data, old, new)
}

// Set stores new into the managed pointer.
func (r *Ref) Set(new unsafe.Pointer) {
	atomic.StorePointer(&r.data, new)
}
