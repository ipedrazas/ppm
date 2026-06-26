package cmd

import (
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold the memory root (index/preferences/glossary + projects/)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		if err := st.Init(); err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Initialized memory at " + st.Root(),
			Data:    map[string]string{"root": st.Root()},
		})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
