package cmd

import (
	"fmt"
	"strings"

	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Full-text search across all memory, with provenance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		hits, err := st.Search(args[0])
		if err != nil {
			return err
		}
		var b strings.Builder
		if len(hits) == 0 {
			fmt.Fprintf(&b, "No matches for %q.", args[0])
		}
		for i, h := range hits {
			if i > 0 {
				b.WriteByte('\n')
			}
			fmt.Fprintf(&b, "- %s: …%s…", h.RelPath, h.Snippet)
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: b.String(),
			Data:    map[string]any{"hits": hits},
		})
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
