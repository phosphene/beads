package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/utils"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage task.md files linked to beads",
	Long: `Manage task.md files for project tracking.
The task command synchronizes a task.md file with the beads database.

Tasks in the file should follow GFM checkbox format:
- [ ] bd-123: Task title
- [/] bd-124: In-progress task
- [x] bd-125: Completed task

If a task has no ID, bd task sync will create it in beads and update the file.`,
	Run: func(cmd *cobra.Command, args []string) {
		syncTaskCmd.Run(cmd, args)
	},
}

var syncTaskCmd = &cobra.Command{
	Use:   "sync [file]",
	Short: "Synchronize task.md with beads",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckReadonly("task sync")

		taskFile := "task.md"
		if len(args) > 0 {
			taskFile = args[0]
		}

		// Find the true path
		if _, err := os.Stat(taskFile); os.IsNotExist(err) {
			// Try .beads/task.md
			altPath := filepath.Join(".beads", "task.md")
			if _, err := os.Stat(altPath); err == nil {
				taskFile = altPath
			} else {
				FatalError("task file %s not found", taskFile)
			}
		}

		ctx := rootCtx
		s := getStore()
		actor := getActorWithGit()

		tasks, err := utils.ParseTaskFile(taskFile)
		if err != nil {
			FatalError("failed to parse task file: %v", err)
		}

		updatedCount := 0
		createdCount := 0

		for i, task := range tasks {
			if task.ID != "" {
				// Update existing issue
				issue, err := s.GetIssue(ctx, task.ID)
				if err != nil {
					WarnError("failed to get issue %s: %v", task.ID, err)
					continue
				}
				if issue == nil {
					WarnError("issue %s not found in beads", task.ID)
					continue
				}

				if issue.Status != task.Status {
					updates := map[string]interface{}{
						"status": task.Status,
					}
					if err := s.UpdateIssue(ctx, task.ID, updates, actor); err != nil {
						WarnError("failed to update issue %s: %v", task.ID, err)
					} else {
						updatedCount++
					}
				}
			} else {
				// Create new issue
				newIssue := &types.Issue{
					Title:     task.Title,
					Status:    task.Status,
					IssueType: types.TypeTask,
					Priority:  2,
					Assignee:  actor,
				}
				if err := s.CreateIssue(ctx, newIssue, actor); err != nil {
					WarnError("failed to create issue for %s: %v", task.Title, err)
				} else {
					tasks[i].ID = newIssue.ID
					createdCount++
				}
			}
		}

		if err := utils.UpdateTaskFile(taskFile, tasks); err != nil {
			FatalError("failed to update task file: %v", err)
		}

		// Flush commit if needed
		if isEmbeddedMode() && (updatedCount > 0 || createdCount > 0) {
			_, _ = s.CommitPending(ctx, actor)
		}

		if !jsonOutput {
			fmt.Printf("%s Synchronized %s: %d updated, %d created\n",
				ui.RenderPass("✓"), taskFile, updatedCount, createdCount)
		}
	},
}

var initTaskCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize task.md from beads",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := rootCtx
		s := getStore()

		// Get ready work
		filter := types.WorkFilter{
			Status: "open",
		}
		issues, err := s.GetReadyWork(ctx, filter)
		if err != nil {
			FatalError("failed to get ready work: %v", err)
		}

		if len(issues) == 0 {
			fmt.Println("No ready work found to initialize task.md")
			return
		}

		file, err := os.Create("task.md")
		if err != nil {
			FatalError("failed to create task.md: %v", err)
		}
		defer file.Close()

		fmt.Fprintln(file, "# Project Tasks")
		fmt.Fprintln(file)
		for _, issue := range issues {
			statusChar := " "
			if issue.Status == types.StatusInProgress {
				statusChar = "/"
			}
			fmt.Fprintf(file, "- [%s] %s: %s\n", statusChar, issue.ID, issue.Title)
		}

		fmt.Printf("%s Initialized task.md with %d issues\n", ui.RenderPass("✓"), len(issues))
	},
}

func init() {
	taskCmd.AddCommand(syncTaskCmd)
	taskCmd.AddCommand(initTaskCmd)
	rootCmd.AddGroup(&cobra.Group{ID: "antigravity", Title: "Antigravity Workflows:"})
	taskCmd.GroupID = "antigravity"
	rootCmd.AddCommand(taskCmd)
}
