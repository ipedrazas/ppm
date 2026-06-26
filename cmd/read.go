package cmd

import (
	"fmt"
	"strings"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var (
	readType string
	readName string
)

var readCmd = &cobra.Command{
	Use:   "read [project]",
	Short: "Read full entry content",
	Long: "Read full content. Omit the project for the workspace index; give a\n" +
		"project but no --type for its index; for collection types pass --name.",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		project := ""
		if len(args) > 0 {
			project = args[0]
		}
		content, err := st.Read(project, memory.EntryType(readType), readName)
		if err != nil {
			if memory.IsMemoryError(err) {
				return err
			}
			return fmt.Errorf("not found: %s", strings.Join(nonEmpty(project, readType, readName), "/"))
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: content,
			Data:    map[string]string{"content": content},
		})
	},
}

func nonEmpty(parts ...string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func init() {
	readCmd.Flags().StringVar(&readType, "type", "", "entry type (summary|focus|decision|question|task|note|conversation)")
	readCmd.Flags().StringVar(&readName, "name", "", "entry name for collection types")
	rootCmd.AddCommand(readCmd)
}
