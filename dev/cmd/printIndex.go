package main

import (
	"fmt"
	"os"

	"github.com/otard95/pass-env/state"
)

func main() {
	if !state.IsInitialized() {
		fmt.Println("pass-env is not initialized. Run 'pass-env init' first.")
		os.Exit(1)
	}

	index := state.GetIndex()

	if len(index) == 0 {
		fmt.Println("Index is empty (no entries)")
		return
	}

	fmt.Printf("Index contents (%d entries):\n\n", len(index))
	for passName, deps := range index {
		fmt.Printf("Pass Name: %s\n", passName)
		depsSlice := deps.Items()
		if len(depsSlice) == 0 {
			fmt.Println("  No cache entries")
		} else {
			fmt.Printf("  Cache entries (%d):\n", len(depsSlice))
			for _, entry := range depsSlice {
				fmt.Printf("    - %s\n", entry)
			}
		}
		fmt.Println()
	}
}
