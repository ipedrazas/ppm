package cmd

import (
	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Manage the project summary (stable overview)",
}

var summarySetCmd = &cobra.Command{
	Use:   "set <project>",
	Short: "Replace the project summary",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setSingleton(cmd, args[0], memory.TypeSummary)
	},
}

var focusCmd = &cobra.Command{
	Use:   "focus",
	Short: "Manage the project focus (volatile current thread)",
}

var focusSetCmd = &cobra.Command{
	Use:   "set <project>",
	Short: "Replace the project focus",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setSingleton(cmd, args[0], memory.TypeFocus)
	},
}

func init() {
	registerContent(summarySetCmd)
	summaryCmd.AddCommand(summarySetCmd)
	rootCmd.AddCommand(summaryCmd)

	registerContent(focusSetCmd)
	focusCmd.AddCommand(focusSetCmd)
	rootCmd.AddCommand(focusCmd)
}
