//go:build cgo && integration

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_WalkthroughGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupCLITestDB(t)

	// 1. Create and close an issue
	out := runBDInProcess(t, tmpDir, "create", "Done issue", "-p", "1", "--json")
	var issue map[string]interface{}
	json.Unmarshal([]byte(out), &issue)
	id := issue["id"].(string)

	runBDInProcess(t, tmpDir, "close", id, "--reason", "Test completed")

	// 2. Run bd walkthrough generate
	runBDInProcess(t, tmpDir, "walkthrough", "generate")

	// 3. Verify walkthrough.md
	walkPath := filepath.Join(tmpDir, "walkthrough.md")
	if _, err := os.Stat(walkPath); os.IsNotExist(err) {
		t.Fatal("walkthrough.md not created")
	}

	content, _ := os.ReadFile(walkPath)
	if !strings.Contains(string(content), "Done issue") {
		t.Errorf("Expected 'Done issue' in walkthrough.md, got: %s", string(content))
	}
	if !strings.Contains(string(content), "Test completed") {
		t.Errorf("Expected 'Test completed' in walkthrough.md, got: %s", string(content))
	}
}
