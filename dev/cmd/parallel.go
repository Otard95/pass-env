package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <pass-name> [pass-name...]\n", os.Args[0])
		os.Exit(1)
	}

	passNames := os.Args[1:]
	var wg sync.WaitGroup
	results := make(chan string, len(passNames))

	fmt.Printf("Starting %d parallel 'pass show' operations...\n\n", len(passNames))

	for _, passName := range passNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			fmt.Printf("[%s] Starting 'pass show'...\n", name)
			cmd := exec.Command("pass", "show", name)
			output, err := cmd.CombinedOutput()

			if err != nil {
				fmt.Printf("[%s] Failed\n", name)
				results <- fmt.Sprintf("❌ %s: %v\n%s", name, err, string(output))
			} else {
				fmt.Printf("[%s] Success (%d bytes)\n", name, len(output))
				results <- fmt.Sprintf("✅ %s: Retrieved successfully (%d bytes)", name, len(output))
			}
		}(passName)
	}

	fmt.Println("Waiting")
	wg.Wait()
	close(results)

	fmt.Println("\nResults:")
	fmt.Println("========")
	for result := range results {
		fmt.Println(result)
	}
}
