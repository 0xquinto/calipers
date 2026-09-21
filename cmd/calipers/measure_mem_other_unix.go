//go:build unix && !darwin && !linux

package main

import (
	"os/exec"
	"syscall"
)

func excelPIDs() []int {
	return nil
}

// processPeakBytes has no portable live source on the BSDs and other
// unix targets: /proc is Linux-only. The peak still gets recorded from
// ru_maxrss after the child exits (rusagePeakBytes); the 25 ms sampler
// just contributes nothing while the process runs.
func processPeakBytes(pid int) int64 {
	return 0
}

// ru_maxrss is in kilobytes on the BSDs, same as Linux.
func rusagePeakBytes(cmd *exec.Cmd) int64 {
	ru, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok || ru == nil {
		return 0
	}
	return ru.Maxrss * 1024
}
