package cmd

import (
	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/spf13/cobra"
)

var (
	taskRef string
	taskURL string
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage task references (rationale only — never live status)",
}

var taskAddCmd = &cobra.Command{
	Use:   "add <project>",
	Short: "Add a task reference + rationale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		extra := []memory.KV{{Key: "ref", Val: taskRef}}
		if taskURL != "" {
			extra = append(extra, memory.KV{Key: "url", Val: taskURL})
		}
		// Default the entry slug to the tracker ref (e.g. ENG-123) when --name
		// is not given, matching the spec's tasks/ENG-123.md convention.
		name := flagName
		if name == "" {
			name = taskRef
		}
		return addCollection(cmd, args[0], memory.TypeTask, name, extra)
	},
}

func init() {
	registerContent(taskAddCmd)
	registerName(taskAddCmd)
	taskAddCmd.Flags().StringVar(&taskRef, "ref", "", "tracker reference, e.g. ENG-123 (required)")
	taskAddCmd.Flags().StringVar(&taskURL, "url", "", "tracker URL")
	_ = taskAddCmd.MarkFlagRequired("ref")
	taskCmd.AddCommand(taskAddCmd)
	addListCommand(taskCmd, memory.TypeTask, false)
	rootCmd.AddCommand(taskCmd)
}
