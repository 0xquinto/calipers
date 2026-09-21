//go:build darwin

package main

import (
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// XNU's proc_info syscall takes a call number before the proc_pidinfo
// arguments; libproc's proc_pidinfo(pid, flavor, arg, buf, size) is a
// wrapper over it. Constants from <sys/proc_info.h>; SYS_PROC_INFO is 336
// on darwin/amd64 and darwin/arm64 in golang.org/x/sys/unix.
const (
	procInfoCallPidinfo = 0x2 // PROC_INFO_CALL_PIDINFO
	procPidTaskInfo     = 4   // PROC_PIDTASKINFO
)

// procTaskInfo mirrors struct proc_taskinfo (96 bytes on 64-bit darwin).
// Only ResidentSize is read; the remaining fields exist for layout
// correctness.
type procTaskInfo struct {
	VirtualSize      uint64
	ResidentSize     uint64
	TotalUser        uint64
	TotalSystem      uint64
	ThreadsUser      uint64
	ThreadsSystem    uint64
	Policy           int32
	Faults           int32
	Pageins          int32
	CowFaults        int32
	MessagesSent     int32
	MessagesReceived int32
	SyscallsMach     int32
	SyscallsUnix     int32
	Csw              int32
	Threadnum        int32
	Numrunning       int32
	Priority         int32
}

func excelPIDs() []int { return nil }

// processPeakBytes returns the task's current resident set size in bytes.
// Darwin has no VmHWM; the 25 ms sampler in measure.go keeps the max.
func processPeakBytes(pid int) int64 {
	if pid <= 0 {
		return 0
	}
	var info procTaskInfo
	n, _, errno := unix.Syscall6(
		unix.SYS_PROC_INFO,
		procInfoCallPidinfo,
		uintptr(pid),
		procPidTaskInfo,
		0,
		uintptr(unsafe.Pointer(&info)),
		unsafe.Sizeof(info),
	)
	if errno != 0 || n < unsafe.Sizeof(info) || info.ResidentSize == 0 {
		return 0
	}
	return int64(info.ResidentSize)
}

// Darwin reports ru_maxrss in bytes; Linux and the BSDs in kilobytes.
func rusagePeakBytes(cmd *exec.Cmd) int64 {
	ru, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok || ru == nil {
		return 0
	}
	return ru.Maxrss
}
