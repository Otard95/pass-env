package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/otard95/pass-env/config"
	"github.com/otard95/pass-env/state"
	"github.com/spf13/cobra"
)

var Delete bool

// aliasCmd represents the alias command
var aliasCmd = &cobra.Command{
	Use:   "alias ALIAS [NAME=PASS_NAME...]",
	Short: "Alias a set for NAME=PASS_NAME pairs",
	Long: `If you where to 'pass-env alias ghp GITHUB_TOKEN=github/token',
the following two commands would be equivalent:
  - 'pass-env GITHUB_TOKEN=github/token gh pr view -c'
  - 'pass-env ghp gh pr view -c'
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		pairs := args[1:]

		if Delete {
			if _, exists := config.Alieses[alias]; exists {
				delete(config.Alieses, alias)
				config.Save()
			}
			return
		}

		if strings.ContainsAny(alias, ": ") {
			fmt.Println("The alias may not contain either ':' or whitespace")
			os.Exit(1)
		}

		for _, pair := range pairs {
			if !state.IsEnvPair(pair) {
				fmt.Printf("Invalid env pair %s\n", pair)
				os.Exit(1)
			}
		}

		config.Alieses[alias] = strings.Join(pairs, " ")
		config.Save()
	},
}

func init() {
	aliasCmd.Flags().BoolVarP(&Delete, "delete", "d", false, "Delete the given alias")
	rootCmd.AddCommand(aliasCmd)
}
