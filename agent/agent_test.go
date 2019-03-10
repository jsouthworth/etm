package agent

import (
	"testing"
	"time"
)

func TestOne(t *testing.T) {
	agt := New(1)
	a := agt.Deref()
	if a != 1 {
		t.Fatalf("got %v, wanted %v\n", a, 1)
	}
	agt.Send(func(cur int) int {
		return 2
	})
	for {
		b := agt.Deref()
		if b == 2 {
			break
		}
	}
}

func TestMany(t *testing.T) {
	agt := New(0)
	expected := 0
	for i := 0; i < 100; i++ {
		expected += i
		go func(i int) {
			agt.Send(func(cur int, next int) int {
				return cur + next
			}, i)
		}(i)
	}
	done := make(chan struct{})
	go func() {
		for {
			got := agt.Deref()
			if got == expected {
				break
			}
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout while waiting for agent to finish")
	}
}
