package cmd

import (
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var waiveCmd = &cobra.Command{
	Use:   "waive <concern-id> <project>",
	Short: "Record a reasoned exception of a standard for a project",
	Long: "Waive a concern (a standard id) for one project so audit reports it as\n" +
		"'waived' with your reason instead of 'fail'. A reason is required — that is\n" +
		"the point: an exception you can't justify shouldn't be silently green.",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		reason, err := contentInput.resolve()
		if err != nil {
			return err
		}
		entry, err := st.Waive(args[1], args[0], reason)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Waived \"" + args[0] + "\" for \"" + args[1] + "\".",
			Data:    entry,
		})
	},
}

func init() {
	registerContent(waiveCmd)
	rootCmd.AddCommand(waiveCmd)
}
