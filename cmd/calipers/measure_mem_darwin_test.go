//go:build darwin

package main

import (
	"testing"
	"unsafe"
)

// The sampler reads ResidentSize out of a hand-defined kernel struct.
// Lock size and offset: a reordered field can still size to 96 while
// reporting pti_virtual_size as RSS.
func TestProcTaskInfoLayout(t *testing.T) {
	if n := unsafe.Sizeof(procTaskInfo{}); n != 96 {
		t.Fatalf("sizeof procTaskInfo = %d, want 96", n)
	}
	if off := unsafe.Offsetof(procTaskInfo{}.ResidentSize); off != 8 {
		t.Fatalf("offsetof ResidentSize = %d, want 8", off)
	}
}
