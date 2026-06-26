// Package cmd wires the Cobra command tree for the ppm memory CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/ipedrazas/ppm/internal/config"
	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

// version is overridable at build time via -ldflags "-X .../cmd.version=v1.2.3".
var version = "dev"

var (
	flagRoot   string
	flagOutput string
	flagPretty bool
)

var rootCmd = &cobra.Command{
	Use:   "ppm",
	Short: "Manage the PM/Product-Owner agent's markdown memory",
	Long: "ppm manages the directory-per-project memory system: typed entries\n" +
		"(decisions, questions, tasks, notes, conversations) plus per-project\n" +
		"index/summary/focus singletons. Output is JSON by default for agent use;\n" +
		"pass -o text (or --pretty) for human-readable output.",
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if flagOutput != string(output.JSON) && flagOutput != string(output.Text) {
			return fmt.Errorf("invalid --output %q: want json or text", flagOutput)
		}
		return nil
	},
}

// Execute runs the root command and returns a process exit code. JSON errors go
// to stdout (uniform machine parsing); human-readable errors go to stderr.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		if format() == output.Text {
			output.RenderError(os.Stderr, output.Text, err)
		} else {
			output.RenderError(os.Stdout, output.JSON, err)
		}
		return 1
	}
	return 0
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&flagRoot, "root", "", "memory root dir (else $"+config.EnvRoot+", else walk-up, else ./memory)")
	pf.StringVarP(&flagOutput, "output", "o", string(output.JSON), "output format: json|text")
	pf.BoolVar(&flagPretty, "pretty", false, "shorthand for --output text")
}

// format resolves the effective output format.
func format() output.Format {
	if flagPretty || flagOutput == string(output.Text) {
		return output.Text
	}
	return output.JSON
}

// openStore resolves the memory root and returns a Store over it.
func openStore() (*memory.Store, error) {
	root, err := config.ResolveRoot(flagRoot)
	if err != nil {
		return nil, err
	}
	return memory.NewStore(root), nil
}

// emit renders a successful result on the command's stdout.
func emit(cmd *cobra.Command, r output.Result) error {
	return output.Render(cmd.OutOrStdout(), format(), r)
}
