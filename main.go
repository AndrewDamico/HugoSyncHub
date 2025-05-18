package main

import (
	"fmt"
	"os"

	"github.com/AndrewDamico/HugoSyncHub/hugoops"
)

func main() {
	// Determine the repo root (current working directory)
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Run the initializer
	if err := hugoops.InitializeSite(root); err != nil {
		fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Hugo site restructured and mounts applied successfully!")
}
