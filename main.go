package hugosynchub

import (
	"fmt"
	"os"

	"github.com/AndrewDamico/HugoSyncHub/hugoops"
)

func main() {
	// 1) Determine your repo root (here: current working directory)
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working dir: %v\n", err)
		os.Exit(1)
	}

	// 2) Call InitializeSite on that root
	if err := hugoops.InitializeSite(root); err != nil {
		fmt.Fprintf(os.Stderr, "initializer error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Hugo site restructured and mounts applied successfully!")
}
