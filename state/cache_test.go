package state

import (
	"os"
	"os/exec"
	"testing"
)

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pass-env-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set Path to temp dir
	oldPath := Path
	Path = tmpDir

	// Get GPG key for testing - use user's own key
	cmd := exec.Command("gpg", "--list-secret-keys", "--keyid-format", "LONG")
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		os.RemoveAll(tmpDir)
		t.Skip("No GPG keys available for testing")
	}

	// Use first available key (simpler approach - just use user's email)
	userEmail := os.Getenv("USER") + "@localhost"

	// Initialize the pass-env store
	cmd = exec.Command("pass", "init", userEmail)
	cmd.Env = append(os.Environ(), "PASSWORD_STORE_DIR="+Store())
	if err := cmd.Run(); err != nil {
		// Try with just getting any GPG ID
		cmd = exec.Command("gpg", "--list-secret-keys", "--with-colons")
		out, _ := cmd.Output()
		if len(out) == 0 {
			os.RemoveAll(tmpDir)
			t.Skip("Could not initialize pass store")
		}
		// Retry with generic init
		cmd = exec.Command("pass", "init", userEmail)
		cmd.Env = append(os.Environ(), "PASSWORD_STORE_DIR="+Store())
		if err := cmd.Run(); err != nil {
			os.RemoveAll(tmpDir)
			t.Skipf("Failed to init pass store: %v", err)
		}
	}

	// Initialize index
	index = make(passNameDependents)

	// Cleanup function
	cleanup := func() {
		Path = oldPath
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestGetCacheMiss(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	cached, hit := GetCache("nonexistent-hash")
	if hit {
		t.Errorf("Expected cache miss, got hit with data: %v", cached)
	}
	if cached != nil {
		t.Errorf("Expected nil for cache miss, got: %v", cached)
	}
}

func TestSetAndGetCache(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	hash := "test-hash-123"
	testData := map[string]string{
		"API_KEY":     "secret-value",
		"DB_PASSWORD": "super-secret",
		"TOKEN":       "auth-token",
	}

	// Set cache
	err := SetCache(hash, testData)
	if err != nil {
		t.Fatalf("SetCache failed: %v", err)
	}

	// Get cache
	cached, hit := GetCache(hash)
	if !hit {
		t.Logf("Store path: %s", Store())
		t.Fatal("Expected cache hit, got miss")
	}

	// Verify data
	if len(cached) != len(testData) {
		t.Errorf("Length mismatch: expected %d, got %d", len(testData), len(cached))
	}

	for key, expectedValue := range testData {
		actualValue, exists := cached[key]
		if !exists {
			t.Errorf("Key '%s' missing from cached data", key)
		} else if actualValue != expectedValue {
			t.Errorf("Key '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
		}
	}
}

func TestUpdateIndex(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	hash := "test-hash-456"
	passNames := []string{"services/api-key", "prod/db", "github/token"}

	// Update index
	err := UpdateIndex(hash, passNames)
	if err != nil {
		t.Fatalf("UpdateIndex failed: %v", err)
	}

	// Verify index
	dependents := GetDependents(passNames...)
	found := false
	for _, dep := range dependents {
		if dep == hash {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Hash '%s' not found in dependents: %v", hash, dependents)
	}
}
