package cmd

import (
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var verdictStatus string

var verdictCmd = &cobra.Command{
	Use:   "verdict <standard-id> <project>",
	Short: "Record the judgement of a manual standard for a project",
	Long: "Resolve a 'manual' standard for one project by recording a pass/fail\n" +
		"judgement with rationale, so audit stops reporting it as 'unknown'. The\n" +
		"agent (or you) judges the semantic standards a built-in check can't.",
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
		entry, err := st.RecordVerdict(args[1], args[0], verdictStatus, reason)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Recorded " + verdictStatus + " verdict for \"" + args[0] + "\" on \"" + args[1] + "\".",
			Data:    entry,
		})
	},
}

func init() {
	verdictCmd.Flags().StringVar(&verdictStatus, "status", "", "judgement: pass|fail (required)")
	_ = verdictCmd.MarkFlagRequired("status")
	registerContent(verdictCmd)
	rootCmd.AddCommand(verdictCmd)
}
