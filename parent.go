package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func run() {
	fmt.Printf("Parent process started (Host PID: %d)\n", os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Start())
	
	fmt.Printf("Container process spawned (Host PID: %d)\n", cmd.Process.Pid)
	setupCgroups(cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Container exited with error: %v\n", err)
	}

	fmt.Println("Cleaning up container resources...")
	os.RemoveAll("./overlay-tmp")
}