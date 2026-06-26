package cmd

import (
	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/spf13/cobra"
)

var conversationCmd = &cobra.Command{
	Use:     "conversation",
	Aliases: []string{"conv"},
	Short:   "Manage conversations that give context to a project",
}

var conversationAddCmd = &cobra.Command{
	Use:   "add <project>",
	Short: "Add a conversation with context to project decisions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addCollection(cmd, args[0], memory.TypeConversation, flagName, nil)
	},
}

func init() {
	registerContent(conversationAddCmd)
	registerName(conversationAddCmd)
	conversationCmd.AddCommand(conversationAddCmd)
	addListCommand(conversationCmd, memory.TypeConversation, false)
	rootCmd.AddCommand(conversationCmd)
}
