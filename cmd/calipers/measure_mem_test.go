package main

import (
	"os"
	"runtime"
	"testing"
)

func TestProcessPeakBytes(t *testing.T) {
	if got := processPeakBytes(0); got != 0 {
		t.Fatalf("pid 0 = %d, want 0", got)
	}
	if got := processPeakBytes(-1); got != 0 {
		t.Fatalf("pid -1 = %d, want 0", got)
	}
	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		if n := processPeakBytes(os.Getpid()); n < 1<<20 {
			t.Fatalf("self pid peakBytes = %d, want >= 1MiB", n)
		}
	}
}
