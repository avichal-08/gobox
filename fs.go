package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func setupOverlay() string {
	lower := "./alpine-fs"
	overlayBase := "./overlay-tmp" 
	upper := filepath.Join(overlayBase, "upper")
	work := filepath.Join(overlayBase, "work") 
	merged := filepath.Join(overlayBase, "merged")
	
	must(os.MkdirAll(upper, 0755))
	must(os.MkdirAll(work, 0755))
	must(os.MkdirAll(merged, 0755))

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)
	must(syscall.Mount("overlay", merged, "overlay", 0, opts))

	return merged
}

func ensureRootFS() {
	if _, err := os.Stat(rootFS); os.IsNotExist(err) {
		fmt.Printf("Root filesystem not found at %s!\n", rootFS)
		fmt.Println("Please run: curl -o alpine.tar.gz https://dl-cdn.alpinelinux.org/alpine/v3.19/releases/x86_64/alpine-minirootfs-3.19.1-x86_64.tar.gz && mkdir alpine-fs && tar -xzf alpine.tar.gz -C alpine-fs")
		os.Exit(1)
	}
}

func setupFilesystem() {
	ensureRootFS()

	must(syscall.Mount(rootFS, rootFS, "bind", syscall.MS_BIND|syscall.MS_REC, ""))

	putold := filepath.Join(rootFS, ".pivot_root")
	must(os.MkdirAll(putold, 0700))

	must(syscall.PivotRoot(rootFS, putold))

	must(syscall.Chdir("/"))

	putoldInside := "/.pivot_root"
	must(syscall.Unmount(putoldInside, syscall.MNT_DETACH))
	must(os.RemoveAll(putoldInside)) 

	must(syscall.Mount("proc", "proc", "proc", 0, ""))
}