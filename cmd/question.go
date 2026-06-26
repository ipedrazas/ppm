package cmd

import (
	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/spf13/cobra"
)

var questionCmd = &cobra.Command{
	Use:   "question",
	Short: "Manage open questions",
}

var questionAddCmd = &cobra.Command{
	Use:   "add <project>",
	Short: "Record an open question (status: open)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addCollection(cmd, args[0], memory.TypeQuestion, flagName, nil)
	},
}

var questionResolveCmd = &cobra.Command{
	Use:   "resolve <project> <name>",
	Short: "Flip a question's status to resolved (optional new content)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		content, err := contentInput.resolveOptional()
		if err != nil {
			return err
		}
		entry, err := st.Write(args[0], memory.TypeQuestion, content, memory.WriteOpts{
			Name: args[1],
			Mode: memory.ModeResolve,
		})
		if err != nil {
			return err
		}
		return emitWrote(cmd, entry)
	},
}

func init() {
	registerContent(questionAddCmd)
	registerName(questionAddCmd)
	registerContent(questionResolveCmd)
	questionCmd.AddCommand(questionAddCmd, questionResolveCmd)
	addListCommand(questionCmd, memory.TypeQuestion, true)
	rootCmd.AddCommand(questionCmd)
}
