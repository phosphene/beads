package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/steveyegge/beads/internal/types"
)

var (
	// taskRegex matches GFM task list items: - [ ] [ID: ]Title
	// Group 1: status (space, /, x, X)
	// Group 2: optional ID (e.g. "bd-123: ")
	// Group 3: title
	taskRegex = regexp.MustCompile(`^\s*-\s+\[([\s/xX])\]\s+(([a-zA-Z0-9-]+):\s+)?(.*)$`)

	// h1Regex matches # Goal Title
	h1Regex = regexp.MustCompile(`^#\s+(.+)$`)
	// h3Regex matches ### Component Name
	h3Regex = regexp.MustCompile(`^###\s+(.+)$`)
	// fileRegex matches #### [MODIFY/NEW/DELETE] [file basename](file:///path)
	fileRegex = regexp.MustCompile(`^####\s+\[(MODIFY|NEW|DELETE)\]\s+\[(.+)\]\(file://(.+)\)$`)
)

// PlanData represents a structured implementation plan
type PlanData struct {
	Goal       string
	Components []*Component
}

type Component struct {
	Name  string
	Files []*PlannedFile
}

type PlannedFile struct {
	Action   string // MODIFY, NEW, DELETE
	Basename string
	Path     string
}

// ParseImplementationPlan parses an implementation_plan.md file
func ParseImplementationPlan(path string) (*PlanData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	plan := &PlanData{}
	var currentComponent *Component

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if matches := h1Regex.FindStringSubmatch(line); matches != nil {
			if plan.Goal == "" {
				plan.Goal = matches[1]
			}
			continue
		}

		if matches := h3Regex.FindStringSubmatch(line); matches != nil {
			currentComponent = &Component{Name: matches[1]}
			plan.Components = append(plan.Components, currentComponent)
			continue
		}

		if matches := fileRegex.FindStringSubmatch(line); matches != nil {
			if currentComponent != nil {
				currentComponent.Files = append(currentComponent.Files, &PlannedFile{
					Action:   matches[1],
					Basename: matches[2],
					Path:     matches[3],
				})
			}
		}
	}

	return plan, scanner.Err()
}

// TaskItem represents a task parsed from markdown
type TaskItem struct {
	ID     string
	Title  string
	Status types.Status
	Line   int
	Raw    string
}

// ParseTaskFile parses a task.md file and returns a list of TaskItems
func ParseTaskFile(path string) ([]*TaskItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []*TaskItem
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := taskRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		statusChar := matches[1]
		id := matches[3]
		title := strings.TrimSpace(matches[4])

		status := types.StatusOpen
		switch strings.ToLower(statusChar) {
		case "/":
			status = types.StatusInProgress
		case "x":
			status = types.StatusClosed
		}

		tasks = append(tasks, &TaskItem{
			ID:     id,
			Title:  title,
			Status: status,
			Line:   lineNum,
			Raw:    line,
		})
	}

	return tasks, scanner.Err()
}

// UpdateTaskFile rewrites a task.md file with updated info
func UpdateTaskFile(path string, tasks []*TaskItem) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	taskMap := make(map[int]*TaskItem)
	for _, t := range tasks {
		taskMap[t.Line] = t
	}

	for i := range lines {
		if task, ok := taskMap[i+1]; ok {
			statusChar := " "
			switch task.Status {
			case types.StatusInProgress:
				statusChar = "/"
			case types.StatusClosed:
				statusChar = "x"
			}

			idPart := ""
			if task.ID != "" {
				idPart = fmt.Sprintf("%s: ", task.ID)
			}

			// Preserve indentation
			indent := ""
			if idx := strings.Index(lines[i], "-"); idx >= 0 {
				indent = lines[i][:idx]
			}

			lines[i] = fmt.Sprintf("%s- [%s] %s%s", indent, statusChar, idPart, task.Title)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
