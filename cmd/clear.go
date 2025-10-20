package cmd

import (
	"fmt"

	"github.com/otard95/pass-env/state"
	"github.com/spf13/cobra"
)

// clearCmd represents the clear command
var clearCmd = &cobra.Command{
	Use:   "clear PASS_NAME...",
	Short: "Clear cached entries for the specified pass names",
	Long: `Remove cached entries from pass-env's internal store that are associated
with the specified password names from the main password store.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !state.IsInitialized() {
			fmt.Println("pass-env is not initialized. Run 'pass-env init' first.")
			return
		}

		cacheEntries := state.GetDependents(args...)

		if len(cacheEntries) == 0 {
			fmt.Println("No cache entries found for the specified pass names.")
			return
		}

		err := state.Clear(cacheEntries...)
		if err != nil {
			fmt.Printf("Error clearing cache entries: %s\n", err)
			return
		}

		err = state.RemoveFromIndex(args...)
		if err != nil {
			fmt.Printf("Error updating index: %s\n", err)
			return
		}

		fmt.Printf("Cleared %d cache entries for: %v\n", len(cacheEntries), args)
	},
}

func init() {
	rootCmd.AddCommand(clearCmd)
}
