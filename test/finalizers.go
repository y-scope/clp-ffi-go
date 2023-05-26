package test

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

type Finalizer struct {
	ch  chan bool
	msg string
}

// We must split NewFinalizers and AssertFinalizers into two functions to
// ensure that the GC call in assertFinalizers will find our pointers
// unreachable. If we combine these functions the runtime will assume the
// pointers in ptrs are still reachable in the caller (even if the caller
// returns immediately after the function call) and will not GC them (so the
// finalizers will not run).
func NewFinalizer[T any](ptr *T) Finalizer {
	fin := Finalizer{
		make(chan bool, 1),
		fmt.Sprintf("%T", ptr),
	}
	// must capture fin.ch for SetFinalizer lambda
	ch := fin.ch
	runtime.SetFinalizer(ptr, func(_ any) { ch <- true })
	return fin
}

// AssertFinalizers checks that each Finalizer channel has been signalled after
// running a GC cycle.
func AssertFinalizers(t *testing.T, fins ...Finalizer) {
	t.Helper()
	runtime.GC()
	for _, fin := range fins {
		select {
		case <-fin.ch:
		case <-time.After(4 * time.Second):
			t.Fatalf("finalizer did not run for: %s", fin.msg)
		}
	}
}
