package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/otard95/pass-env/lib/fs"
	"github.com/otard95/pass-env/state"
)

type aliases map[string]string

var (
	Alieses       aliases
	NotFoundError = errors.New("Not Found")
)

func Save() {
	aliasesFile, err := getFile()
	if err != nil {
		fmt.Printf("%e", err)
		os.Exit(1)
	}

	dir := path.Dir(aliasesFile)
	err = os.MkdirAll(dir, os.ModeDir|0700)
	if err != nil {
		fmt.Printf("Unable to create config directory: %e", err)
		os.Exit(1)
	}

	lines := make([]string, len(Alieses))
	for k, v := range Alieses {
		lines = append(lines, fmt.Sprintf("%s: %s", k, v))
	}

	err = os.WriteFile(aliasesFile, []byte(strings.Join(lines, "\n")), 0600)
	if err != nil {
		fmt.Printf("Unable to write config file: %e", err)
		os.Exit(1)
	}
}

func getFile() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("Unable to resolve config dir: %e", err)
	}

	return path.Join(configDir, "pass-env", "aliases"), nil
}

func init() {
	defer func() {
		if Alieses == nil {
			Alieses = make(aliases)
		}
	}()

	aliasesFile, err := getFile()
	if err != nil {
		fmt.Printf("%e", err)
		os.Exit(1)
	}

	if !fs.IsFile(aliasesFile) {
		return
	}

	content, err := os.ReadFile(aliasesFile)
	if err != nil {
		fmt.Printf("Unable to read aliases config file: %e\n", err)
		os.Exit(1)
	}

	Alieses = make(aliases, strings.Count(string(content), "\n")+1)

validateLines:
	for line := range strings.Lines(string(content)) {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		parts := strings.SplitN(trimmed, ": ", 2)
		if len(parts) != 2 {
			fmt.Printf("WARN: Invalid config line in '%s':\n  %s\n", aliasesFile, line)
			continue
		}

		for pair := range strings.SplitSeq(parts[1], " ") {
			if !state.IsEnvPair(pair) {
				fmt.Printf("WARN: Invalid env pair '%s' in '%s':\n  %s\n", pair, aliasesFile, line)
				continue validateLines
			}
		}

		Alieses[parts[0]] = parts[1]
	}
}
