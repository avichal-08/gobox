package main

import (
	"fmt"
	"os"
	"syscall"
)

func ensureRootFS() {
	if _, err := os.Stat(rootFS); os.IsNotExist(err) {
		fmt.Printf("Root filesystem not found at %s!\n", rootFS)
		fmt.Println("Please run: curl -o alpine.tar.gz https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.1-x86_64.tar.gz && mkdir alpine-fs && tar -xzf alpine.tar.gz -C alpine-fs")
		os.Exit(1)
	}
}

func setupFilesystem() {
	ensureRootFS()
	
	must(syscall.Chroot(rootFS))
	must(syscall.Chdir("/"))

	must(syscall.Mount("proc", "proc", "proc", 0, ""))
}