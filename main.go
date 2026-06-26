package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gobox [build|run|child] [args]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		fmt.Println("--> Build command triggered")
	case "run":
		run()
	case "child":
		child()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
	}
}

func run() {
	fmt.Printf("Running %v as parent PID %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR running child:", err)
		os.Exit(1)
	}
}

func child() {
	fmt.Printf("Running %v as child PID %d\n", os.Args[2:], os.Getpid())

	syscall.Sethostname([]byte("gobox-container"))

	cmdPath, err := exec.LookPath(os.Args[2])
	if err != nil {
		fmt.Println("Command not found:", err)
		os.Exit(1)
	}

	if err := syscall.Exec(cmdPath, os.Args[2:], os.Environ()); err != nil {
		fmt.Println("ERROR running user command:", err)
		os.Exit(1)
	}
}