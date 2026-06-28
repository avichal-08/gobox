package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func child() {
	fmt.Printf("Container initializing... (Container PID: %d)\n", os.Getpid())

	must(syscall.Sethostname([]byte(containerHostname)))

	setupFilesystem()

	must(os.MkdirAll("/etc", 0755))
	os.Remove("/etc/resolv.conf")
	must(os.WriteFile("/etc/resolv.conf", []byte("nameserver 8.8.8.8\n"), 0644))

	cmdPath, err := exec.LookPath(os.Args[2])
	if err != nil {
		fmt.Printf("Command not found: %s\n", os.Args[2])
		os.Exit(1)
	}

	must(syscall.Exec(cmdPath, os.Args[2:], os.Environ()))
}