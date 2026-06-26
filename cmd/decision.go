package cmd

import (
	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/spf13/cobra"
)

var decisionCmd = &cobra.Command{
	Use:   "decision",
	Short: "Manage decisions",
}

var decisionAddCmd = &cobra.Command{
	Use:   "add <project>",
	Short: "Record a dated, atomic decision + rationale",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addCollection(cmd, args[0], memory.TypeDecision, flagName, nil)
	},
}

func init() {
	registerContent(decisionAddCmd)
	registerName(decisionAddCmd)
	decisionCmd.AddCommand(decisionAddCmd)
	addListCommand(decisionCmd, memory.TypeDecision, false)
	rootCmd.AddCommand(decisionCmd)
}
