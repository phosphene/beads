package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
)

var walkthroughCmd = &cobra.Command{
	Use:   "walkthrough",
	Short: "Manage walkthrough.md files linked to beads",
	Long: `Manage walkthrough.md files for project verification.
The walkthrough command generates a summary of recently completed work.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var generateWalkthroughCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate walkthrough.md from recently closed issues",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := rootCtx
		s := getStore()

		// Get issues closed in the last 24 hours (default)
		since := time.Now().Add(-24 * time.Hour)
		
		filter := types.IssueFilter{
			Status: func() *types.Status { s := types.StatusClosed; return &s }(),
		}
		
		issues, err := s.SearchIssues(ctx, "", filter)
		if err != nil {
			FatalError("failed to get closed issues: %v", err)
		}

		var recentIssues []*types.Issue
		for _, issue := range issues {
			if issue.ClosedAt != nil && issue.ClosedAt.After(since) {
				recentIssues = append(recentIssues, issue)
			}
		}

		if len(recentIssues) == 0 {
			fmt.Println("No recently closed issues found to generate walkthrough.md")
			return
		}

		file, err := os.Create("walkthrough.md")
		if err != nil {
			FatalError("failed to create walkthrough.md: %v", err)
		}
		defer file.Close()

		fmt.Fprintln(file, "# Walkthrough: Summary of Completed Work")
		fmt.Fprintln(file)
		fmt.Fprintln(file, "## Changes Made")
		fmt.Fprintln(file)
		for _, issue := range recentIssues {
			fmt.Fprintf(file, "### [CLOSE] %s: %s\n", issue.ID, issue.Title)
			if issue.CloseReason != "" {
				fmt.Fprintf(file, "- **Reason**: %s\n", issue.CloseReason)
			}
			fmt.Fprintln(file)
		}

		fmt.Fprintln(file, "## Verification Results")
		fmt.Fprintln(file)
		fmt.Fprintln(file, "- [ ] All manual verification steps completed")
		fmt.Fprintln(file, "- [ ] Automated tests passed")

		fmt.Printf("%s Generated walkthrough.md with %d issues\n", ui.RenderPass("✓"), len(recentIssues))
	},
}

func init() {
	walkthroughCmd.AddCommand(generateWalkthroughCmd)
	walkthroughCmd.GroupID = "antigravity"
	rootCmd.AddCommand(walkthroughCmd)
}
