//go:build cgo && integration

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_TaskSync(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupCLITestDB(t)

	// 1. Create a task.md with a new task
	taskContent := "# Tasks\n\n- [ ] New task\n"
	taskPath := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(taskPath, []byte(taskContent), 0644); err != nil {
		t.Fatalf("failed to write task.md: %v", err)
	}

	// 2. Run bd task sync
	runBDInProcess(t, tmpDir, "task", "sync")

	// 3. Verify file was updated with an ID
	newContent, _ := os.ReadFile(taskPath)
	if !strings.Contains(string(newContent), "bd-") {
		t.Errorf("Expected ID in task.md, got: %s", string(newContent))
	}

	// 4. Update status in file and sync back
	updatedContent := strings.Replace(string(newContent), "[ ]", "[/]", 1)
	os.WriteFile(taskPath, []byte(updatedContent), 0644)
	runBDInProcess(t, tmpDir, "task", "sync")

	// 5. Verify status in DB
	id := ""
	lines := strings.Split(updatedContent, "\n")
	for _, line := range lines {
		if strings.Contains(line, "bd-") {
			parts := strings.Split(line, " ")
			for _, p := range parts {
				if strings.HasPrefix(p, "bd-") {
					id = strings.TrimSuffix(p, ":")
					break
				}
			}
		}
	}

	out := runBDInProcess(t, tmpDir, "show", id, "--json")
	if !strings.Contains(out, "in_progress") {
		t.Errorf("Expected status in_progress in DB, got: %s", out)
	}
}

func TestCLI_TaskInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupCLITestDB(t)

	runBDInProcess(t, tmpDir, "create", "Ready item", "-p", "1")
	runBDInProcess(t, tmpDir, "task", "init")

	taskPath := filepath.Join(tmpDir, "task.md")
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		t.Fatal("task.md not created")
	}

	content, _ := os.ReadFile(taskPath)
	if !strings.Contains(string(content), "Ready item") {
		t.Errorf("Expected 'Ready item' in task.md, got: %s", string(content))
	}
}
