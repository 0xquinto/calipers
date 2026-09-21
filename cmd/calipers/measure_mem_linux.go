//go:build linux

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func excelPIDs() []int {
	return nil
}

func processPeakBytes(pid int) int64 {
	if pid <= 0 {
		return 0
	}
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if !bytes.HasPrefix(line, []byte("VmHWM:")) {
			continue
		}
		fields := strings.Fields(string(line))
		if len(fields) < 2 {
			return 0
		}
		kb, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return 0
		}
		return kb * 1024
	}
	return 0
}

// ru_maxrss is in kilobytes on Linux and the BSDs.
func rusagePeakBytes(cmd *exec.Cmd) int64 {
	ru, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok || ru == nil {
		return 0
	}
	return ru.Maxrss * 1024
}
