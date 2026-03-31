package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/utils"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage implementation plans linked to beads",
	Long: `Manage implementation_plan.md files for project planning.
The plan command can import an Antigravity-formatted implementation plan
into beads, creating an Epic for the goal and Tasks for the proposed changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var importPlanCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import implementation_plan.md into beads",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckReadonly("plan import")

		planFile := "implementation_plan.md"
		if len(args) > 0 {
			planFile = args[0]
		}

		ctx := rootCtx
		s := getStore()
		actor := getActorWithGit()

		plan, err := utils.ParseImplementationPlan(planFile)
		if err != nil {
			FatalError("failed to parse implementation plan: %v", err)
		}

		if plan.Goal == "" {
			FatalError("no goal found in implementation plan")
		}

		// 1. Create Epic for the goal
		epic := &types.Issue{
			Title:     plan.Goal,
			IssueType: types.TypeEpic,
			Priority:  1,
			Status:    types.StatusOpen,
			Assignee:  actor,
		}
		if err := s.CreateIssue(ctx, epic, actor); err != nil {
			FatalError("failed to create Epic: %v", err)
		}

		taskCount := 0
		for _, comp := range plan.Components {
			for _, file := range comp.Files {
				// 2. Create Task for each file
				actionStr := strings.ToLower(file.Action)
				title := fmt.Sprintf("%s %s in %s", actionStr, file.Basename, comp.Name)
				task := &types.Issue{
					Title:     title,
					IssueType: types.TypeTask,
					Priority:  2,
					Status:    types.StatusOpen,
					Assignee:  actor,
					Description: fmt.Sprintf("Action: %s\nFile: %s\nPath: %s",
						file.Action, file.Basename, file.Path),
				}
				if err := s.CreateIssue(ctx, task, actor); err != nil {
					WarnError("failed to create Task for %s: %v", file.Basename, err)
					continue
				}

				// 3. Link Task to Epic (Parent-Child)
				dep := &types.Dependency{
					IssueID:     task.ID,
					DependsOnID: epic.ID,
					Type:        types.DepParentChild,
				}
				if err := s.AddDependency(ctx, dep, actor); err != nil {
					WarnError("failed to link task %s to epic %s: %v", task.ID, epic.ID, err)
				}
				taskCount++
			}
		}

		// Flush commit if needed
		if isEmbeddedMode() {
			_, _ = s.CommitPending(ctx, actor)
		}

		if !jsonOutput {
			fmt.Printf("%s Imported %s:\n", ui.RenderPass("✓"), planFile)
			fmt.Printf("  Epic: %s (%s)\n", epic.ID, epic.Title)
			fmt.Printf("  Tasks: %d created and linked\n", taskCount)
		}
	},
}

func init() {
	planCmd.AddCommand(importPlanCmd)
	planCmd.GroupID = "antigravity"
	rootCmd.AddCommand(planCmd)
}
