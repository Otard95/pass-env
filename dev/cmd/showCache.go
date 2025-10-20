package main

import (
	"fmt"
	"os"

	"github.com/otard95/pass-env/state"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <cache-hash>\n", os.Args[0])
		os.Exit(1)
	}

	hash := os.Args[1]

	if !state.IsInitialized() {
		fmt.Println("pass-env is not initialized. Run 'pass-env init' first.")
		os.Exit(1)
	}

	cached, hit := state.GetCache(hash)
	if !hit {
		fmt.Printf("Cache miss for hash: %s\n", hash)
		os.Exit(1)
	}

	fmt.Printf("Cache hit for hash: %s\n", hash)
	fmt.Println("\nCached environment variables:")
	for name, value := range cached {
		fmt.Printf("  %s=%s\n", name, value)
	}
}
