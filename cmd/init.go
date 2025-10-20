package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"slices"

	"github.com/otard95/pass-env/state"
	"github.com/spf13/cobra"
)

var flag_passStore, flag_gpg string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize pass-env",
	Long: `This initializes pass-env's, internal store and state.

	This command will try and find your existing pass-store and its gpg key, if
	fails to do so or you wan't to use a different one, you can use the
	--pass-store and --gpg flags to specify these.`,
	Run: func(cmd *cobra.Command, args []string) {
		if state.IsInitialized() {
			log.Fatalln("pass-env is already initialized")
		}

		passStore, err := requireDir(
			flag_passStore,
			os.Getenv("PASSWORD_STORE_DIR"),
			os.Getenv("HOME")+"/.password-store",
		)
		if err != nil {
			log.Fatalf("Could not find pass store: %s", err)
		}

		gpgKey, err := findGPGKeyInPath(passStore)
		if err != nil {
			log.Fatalf("Could not automatically find gpg key: %s", err)
		}

		err = state.Init(passStore, gpgKey)
		if err != nil {
			log.Fatalf("Failed to initialize: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(
		&flag_passStore, "pass-store", "s", "",
		"Use the pass-store at this path instead of the default in environment variables",
	)
	initCmd.Flags().StringVarP(
		&flag_gpg, "gpg", "g", "",
		"Use this gpg key, instead of the one from the existing pass-store",
	)
}

func requireDir(paths ...string) (string, error) {
	nonEmptyIndex := slices.IndexFunc(paths, func(s string) bool { return s != "" })
	if nonEmptyIndex == -1 {
		return "", errors.New("Required directory not provided")
	}

	stat, err := os.Stat(paths[nonEmptyIndex])
	if err != nil {
		return "", err
	}

	if !stat.IsDir() {
		return "", fmt.Errorf("'%s' is not a directory", paths[nonEmptyIndex])
	}

	return paths[nonEmptyIndex], nil
}

func findGPGKeyInPath(storePath string) (string, error) {
	files, err := os.ReadDir(storePath)
	if err != nil {
		return "", err
	}

	gpgIndex := slices.IndexFunc(
		files, func(file os.DirEntry) bool { return file.Name() == ".gpg-id" },
	)
	if gpgIndex == -1 {
		return "", fmt.Errorf("Unable to find gpg key in '%s'", storePath)
	}

	gpgFilePath := path.Join(storePath, files[gpgIndex].Name())

	content, err := os.ReadFile(gpgFilePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
