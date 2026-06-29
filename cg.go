package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func setupCgroups(pid int) {
	if err := os.MkdirAll(cgroupDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create cgroup directory: %v\n", err)
		return
	}

	must(os.WriteFile(filepath.Join(cgroupDir, "pids.max"), []byte(maxProcesses), 0700))

	must(os.WriteFile(filepath.Join(cgroupDir, "memory.max"), []byte(maxMemory), 0700))

	must(os.WriteFile(filepath.Join(cgroupDir, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700))
}