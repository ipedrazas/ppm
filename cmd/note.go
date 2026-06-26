package cmd

import (
	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage notes (explicit miscellany — not a default bucket)",
}

var noteAddCmd = &cobra.Command{
	Use:   "add <project>",
	Short: "Add a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addCollection(cmd, args[0], memory.TypeNote, flagName, nil)
	},
}

func init() {
	registerContent(noteAddCmd)
	registerName(noteAddCmd)
	noteCmd.AddCommand(noteAddCmd)
	addListCommand(noteCmd, memory.TypeNote, false)
	rootCmd.AddCommand(noteCmd)
}
