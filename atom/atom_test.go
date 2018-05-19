package atom

import (
	"testing"
)

func TestSwap(t *testing.T) {
	a := New(1)
	b := a.Deref()
	if b != 1 {
		t.Fatalf("got %v, wanted %v\n", b, 1)
	}
	a.Swap(func(cur interface{}, args ...interface{}) interface{} {
		return 2
	})
	b = a.Deref()
	if b != 2 {
		t.Fatalf("got %v, wanted %v\n", b, 2)
	}
}

func TestReset(t *testing.T) {
	a := New(1)
	b := a.Deref()
	if b != 1 {
		t.Fatalf("got %v, wanted %v\n", b, 1)
	}
	a.Reset(2)
	b = a.Deref()
	if b != 2 {
		t.Fatalf("got %v, wanted %v\n", b, 2)
	}
}
