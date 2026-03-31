//go:build cgo && integration

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_PlanImport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupCLITestDB(t)

	// 1. Create a sample implementation_plan.md
	planContent := `# Global Goal

## Proposed Changes

### Component A
#### [MODIFY] [file1.go](file:///path/to/file1.go)
#### [NEW] [file2.go](file:///path/to/file2.go)
`
	planPath := filepath.Join(tmpDir, "implementation_plan.md")
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatalf("failed to write plan: %v", err)
	}

	// 2. Run bd plan import
	runBDInProcess(t, tmpDir, "plan", "import")

	// 3. Verify Epic and Tasks were created
	out := runBDInProcess(t, tmpDir, "list", "--json")
	if !strings.Contains(out, "Global Goal") {
		t.Errorf("Expected Epic 'Global Goal' in DB, got: %s", out)
	}
	if !strings.Contains(out, "modify file1.go in Component A") {
		t.Errorf("Expected Task for file1.go in DB, got: %s", out)
	}
	if !strings.Contains(out, "new file2.go in Component A") {
		t.Errorf("Expected Task for file2.go in DB, got: %s", out)
	}
}
