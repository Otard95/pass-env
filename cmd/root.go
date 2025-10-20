package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/otard95/pass-env/state"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "pass-env [OPTIONS] NAME=PASS_NAME... COMMAND [ARG]...",
	Long: `pass-env fetches secrets from your password store and sets them as environment
variables before executing a command.

OPTIONS
    Are the same as env(1), and must come before the envs

EXIT STATUS:
   128    invalid arguments
   129    if secret is not found
   -      the exit status of the env(1) command`,
	Example: `  # Run Rails console with database password
  pass-env DB_PASSWORD=prod/database/password rails console

  # Use multiple secrets
  pass-env TOKEN=github/token SLACK_KEY=slack/webhook ./deploy.sh

  # Pass through env options (-i to ignore inherited environment)
  pass-env -i TOKEN=github/token gh pr view -c`,
	Args:                  cobra.ArbitraryArgs,
	DisableFlagParsing:    true,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		parsed, err := parseArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(128)
		}

		err = validateParsedArgs(parsed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(128)
		}

		cacheKey := generateCacheKey(parsed.EnvPairs)

		var envVars map[string]string
		cached, hit := state.GetCache(cacheKey)
		if hit {
			envVars = cached
		} else {
			secrets, err := state.GetSecrets(parsed.EnvPairs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(129)
			}
			envVars = secrets

			err = state.SetCache(cacheKey, envVars)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to cache secrets: %v\n", err)
			}

			passNames := make([]string, 0, len(parsed.EnvPairs))
			for _, passName := range parsed.EnvPairs {
				passNames = append(passNames, passName)
			}
			err = state.UpdateIndex(cacheKey, passNames)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update index: %v\n", err)
			}
		}

		// Build env command args: [options...] NAME=value... command [args...]
		envArgs := make([]string, 0)
		envArgs = append(envArgs, parsed.EnvOpts...)
		for name, value := range envVars {
			envArgs = append(envArgs, fmt.Sprintf("%s=%s", name, value))
		}
		envArgs = append(envArgs, parsed.Command...)

		execCmd := exec.Command("env", envArgs...)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		err = execCmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type ParsedArgs struct {
	EnvPairs map[string]string
	Command  []string
	EnvOpts  []string
}

func isCliFlag(s string) bool {
	return strings.HasPrefix(s, "-")
}

func isEnvPair(s string) bool {
	if !strings.Contains(s, "=") {
		return false
	}
	parts := strings.SplitN(s, "=", 2)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func parseEnvPair(s string) (name, passName string, err error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid env pair: %s", s)
	}
	if parts[0] == "" {
		return "", "", fmt.Errorf("empty name in env pair: %s", s)
	}
	if parts[1] == "" {
		return "", "", fmt.Errorf("empty pass name in env pair: %s", s)
	}
	return parts[0], parts[1], nil
}

func parseArgs(args []string) (*ParsedArgs, error) {
	parsed := &ParsedArgs{
		EnvPairs: make(map[string]string),
		EnvOpts:  []string{},
		Command:  []string{},
	}

	i := 0
	for ; i < len(args) && isCliFlag(args[i]); i++ {
		parsed.EnvOpts = append(parsed.EnvOpts, args[i])
	}

	for ; i < len(args) && isEnvPair(args[i]); i++ {
		name, passName, err := parseEnvPair(args[i])
		if err != nil {
			return nil, err
		}
		parsed.EnvPairs[name] = passName
	}

	if i < len(args) {
		parsed.Command = args[i:]
	}

	return parsed, nil
}

func validateParsedArgs(parsed *ParsedArgs) error {
	if len(parsed.EnvPairs) == 0 {
		return fmt.Errorf("no NAME=PASS_NAME pairs provided")
	}
	if len(parsed.Command) == 0 {
		return fmt.Errorf("no command provided")
	}
	return nil
}

func generateCacheKey(envPairs map[string]string) string {
	pairs := make([]string, 0, len(envPairs))
	for name, passName := range envPairs {
		pairs = append(pairs, fmt.Sprintf("%s=%s", name, passName))
	}
	sort.Strings(pairs)

	hash := sha256.New()
	for _, pair := range pairs {
		hash.Write([]byte(pair))
		hash.Write([]byte("\n"))
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}
