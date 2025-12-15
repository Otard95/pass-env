package state

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/otard95/pass-env/lib/fs"
	"github.com/otard95/pass-env/lib/set"
)

// The cache entries in pass-env's own store that are dependent on pass names
type passNameDependents map[string]set.Set[string]

var Path string
var index passNameDependents

func Store() string {
	return path.Join(Path, "store")
}

func PassStore() string {
	return path.Join(Path, ".pass-store")
}

func StoreIndex() string {
	return path.Join(Path, "store.index")
}

func IsInitialized() bool {
	return fs.IsDir(Path) && fs.IsDir(Store()) && fs.IsFile(StoreIndex()) && fs.IsLink(PassStore())
}

func Init(passStore, gpgKey string) error {
	err := os.MkdirAll(Path, os.ModeDir|0600)
	if err != nil {
		return fmt.Errorf("Failed to create directory '%s': %s", Path, err)
	}

	passCmd := exec.Command("pass", "init", gpgKey)
	passCmd.Env = append(passCmd.Env, fmt.Sprintf("PASSWORD_STORE_DIR=%s", Store()))

	out, err := passCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("'pass init' failed: %s\n%s", err, out)
	}

	err = os.Symlink(passStore, PassStore())
	if err != nil {
		return fmt.Errorf("Failed to create link ('%s') to pass store '%s': %s", PassStore(), passStore, err)
	}

	err = WriteIndex()
	if err != nil {
		return err
	}

	return nil
}

func WriteIndex() error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(index)
	if err != nil {
		return fmt.Errorf("Failed encode index: %s", err)
	}

	err = os.WriteFile(StoreIndex(), buffer.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("Failed to write index '%s': %s", StoreIndex(), err)
	}

	return nil
}

func Index() passNameDependents {
	return index
}

func GetDependents(names ...string) []string {
	result := make(set.Set[string])

	for _, name := range names {
		if deps, exists := index[name]; exists {
			result.Merge(deps)
		}
	}

	return result.Items()
}

func Clear(cacheEntries ...string) error {
	var errors []error

	for _, entry := range cacheEntries {
		passCmd := exec.Command("pass", "rm", "-f", entry)
		passCmd.Env = append(os.Environ(), fmt.Sprintf("PASSWORD_STORE_DIR=%s", Store()))

		out, err := passCmd.CombinedOutput()
		if err != nil {
			if !strings.HasSuffix(string(out), "is not in the password store.") {
				errors = append(errors, fmt.Errorf("failed to remove '%s': %s\n%s", entry, err, out))
			}
		}
	}

	if len(errors) > 0 {
		var errMsg string
		for _, e := range errors {
			errMsg += e.Error() + "\n"
		}
		return fmt.Errorf("errors during clear:\n%s", errMsg)
	}

	return nil
}

func RemoveFromIndex(names ...string) error {
	for _, name := range names {
		delete(index, name)
	}

	return WriteIndex()
}

// GetCache retrieves cached environment variables for a given hash.
// Returns the cached env vars as a map (NAME -> value) and a boolean indicating cache hit.
func GetCache(hash string) (map[string]string, bool) {
	passCmd := exec.Command("pass", "show", hash)
	passCmd.Env = append(os.Environ(), fmt.Sprintf("PASSWORD_STORE_DIR=%s", Store()))

	out, err := passCmd.CombinedOutput()
	if err != nil {
		return nil, false
	}

	envVars := make(map[string]string)
	decoder := gob.NewDecoder(bytes.NewBuffer(out))
	err = decoder.Decode(&envVars)
	if err != nil {
		return nil, false
	}

	return envVars, true
}

func SetCache(hash string, envVars map[string]string) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(envVars)
	if err != nil {
		return fmt.Errorf("failed to encode cache data: %s", err)
	}

	passCmd := exec.Command("pass", "insert", "-m", "-f", hash)
	passCmd.Env = append(os.Environ(), fmt.Sprintf("PASSWORD_STORE_DIR=%s", Store()))
	passCmd.Stdin = &buffer

	out, err := passCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to store cache entry '%s': %s\n%s", hash, err, out)
	}

	return nil
}

// UpdateIndex adds pass names to the index, mapping them to a cache hash
func UpdateIndex(hash string, passNames []string) error {
	if index == nil {
		index = make(passNameDependents)
	}

	for _, passName := range passNames {
		if _, exists := index[passName]; !exists {
			index[passName] = make(set.Set[string])
		}
		s := index[passName]
		s.Add(hash)
		index[passName] = s
	}

	return WriteIndex()
}

// GetSecrets fetches secrets from the parent pass store in parallel.
// Returns a map of NAME -> secret value (first line only).
func GetSecrets(envPairs map[string]string) (map[string]string, error) {
	type result struct {
		name  string
		value string
		err   error
	}

	results := make(chan result, len(envPairs))
	var wg sync.WaitGroup

	for name, passName := range envPairs {
		wg.Add(1)
		go func(envName, passPath string) {
			defer wg.Done()

			passCmd := exec.Command("pass", "show", passPath)
			passCmd.Env = append(os.Environ(), fmt.Sprintf("PASSWORD_STORE_DIR=%s", PassStore()))

			out, err := passCmd.CombinedOutput()
			if err != nil {
				results <- result{
					name: envName,
					err:  fmt.Errorf("secret '%s' not found in password store", passPath),
				}
				return
			}

			lines := strings.Split(string(out), "\n")
			secretValue := strings.TrimSpace(lines[0])

			results <- result{
				name:  envName,
				value: secretValue,
			}
		}(name, passName)
	}

	wg.Wait()
	close(results)

	secrets := make(map[string]string)
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		secrets[res.name] = res.value
	}

	return secrets, nil
}

func IsEnvPair(s string) bool {
	if !strings.Contains(s, "=") {
		return false
	}
	parts := strings.SplitN(s, "=", 2)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func init() {
	// Allow overriding path for testing
	Path = os.Getenv("PASS_ENV_STATE_DIR")
	if Path == "" {
		home := os.Getenv("HOME")
		if home == "" {
			log.Fatal("HOME environment variable not set.")
		}
		Path = path.Join(home, ".local", "share", "pass-env")
	}

	if IsInitialized() {
		data, err := os.ReadFile(StoreIndex())
		if err == nil && len(data) > 0 {
			decoder := gob.NewDecoder(bytes.NewBuffer(data))
			err = decoder.Decode(&index)
			if err != nil {
				fmt.Printf(
					"The index file is corrupt, this is not a big issue, you just wont be able to clear cache based on pass-name's\n%s",
					err,
				)
			}
		}
	}
}
