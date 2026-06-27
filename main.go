package main

import (
	"fmt"
	"os"
)

// main simply acts as our CLI router
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gobox [run|child] <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}