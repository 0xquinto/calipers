package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestMeasureChildCapturesTransientPeak(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the spike helper needs mmap; Windows already samples the kernel-maintained PeakWorkingSetSize")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("cc not available")
	}
	// No sampler ticks during the child's life: the spike is invisible to
	// the sampler and only ru_maxrss can report it, so the red side of
	// this test is deterministic instead of a race against the 25 ms tick.
	old := measureSampleEvery
	measureSampleEvery = time.Hour
	t.Cleanup(func() { measureSampleEvery = old })
	dir := t.TempDir()
	src := filepath.Join(dir, "spike.c")
	bin := filepath.Join(dir, "spike")
	// The strided read into a volatile sink keeps the mapping live under
	// -O2. mmap/munmap instead of malloc/free: darwin keeps freed malloc
	// pages in the task's RSS, while munmap drops it deterministically.
	code := `#include <string.h>
#include <sys/mman.h>

volatile unsigned long sink;

int main(void) {
    size_t n = 64UL << 20;
    char *p = mmap(0, n, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    if (p == MAP_FAILED) return 1;
    memset(p, 1, n);
    unsigned long s = 0;
    for (size_t i = 0; i < n; i += 4096) s += (unsigned char)p[i];
    sink = s;
    munmap(p, n);
    return 0;
}
`
	if err := os.WriteFile(src, []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-O2", "-o", bin, src).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	sample, err := measureChild(bin, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(48 << 20); sample.peakBytes < want {
		t.Fatalf("peakBytes = %d, want >= %d (64 MiB transient spike)", sample.peakBytes, want)
	}
}
