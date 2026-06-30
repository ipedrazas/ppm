package cmd

import (
	"fmt"
	"strings"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var (
	stdTitle     string
	stdAppliesTo string
	stdSeverity  string
	stdCheck     string
)

var standardCmd = &cobra.Command{
	Use:   "standard",
	Short: "Manage cross-cutting standards (workspace-level invariants)",
}

var standardAddCmd = &cobra.Command{
	Use:   "add <id>",
	Short: "Declare a standard: a check bound to a scope and severity",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		body, err := contentInput.resolveOptional()
		if err != nil {
			return err
		}
		std, err := st.AddStandard(args[0], stdTitle, stdAppliesTo, stdSeverity, stdCheck, body)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Declared standard \"" + std.ID + "\".",
			Data:    std,
		})
	},
}

var standardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all standards",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		stds, err := st.ListStandards()
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatStandards(stds),
			Data:    map[string]any{"standards": stds},
		})
	},
}

var standardShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a standard's full definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		std, err := st.ReadStandard(args[0])
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatStandard(std),
			Data:    std,
		})
	},
}

var standardRetireCmd = &cobra.Command{
	Use:   "retire <id>",
	Short: "Retire a standard (kept for history, skipped by audit)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		std, err := st.RetireStandard(args[0])
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Retired standard \"" + std.ID + "\".",
			Data:    std,
		})
	},
}

func formatStandards(stds []memory.Standard) string {
	if len(stds) == 0 {
		return "No standards defined."
	}
	var b strings.Builder
	b.WriteString("Standards:")
	for _, s := range stds {
		b.WriteByte('\n')
		fmt.Fprintf(&b, "- %s [%s/%s] %s → %s", s.ID, s.Status, s.Severity, s.Check, s.AppliesTo)
		if s.Title != "" {
			fmt.Fprintf(&b, ": %s", s.Title)
		}
	}
	return b.String()
}

func formatStandard(s *memory.Standard) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%s)\n", s.ID, s.Status)
	fmt.Fprintf(&b, "title: %s\n", s.Title)
	fmt.Fprintf(&b, "applies-to: %s · severity: %s · check: %s", s.AppliesTo, s.Severity, s.Check)
	if s.Body != "" {
		fmt.Fprintf(&b, "\n\n%s", s.Body)
	}
	return b.String()
}

func init() {
	af := standardAddCmd.Flags()
	af.StringVar(&stdTitle, "title", "", "human-readable standard title")
	af.StringVar(&stdAppliesTo, "applies-to", "all", "scope: all | tag:<t> | comma-separated slugs")
	af.StringVar(&stdSeverity, "severity", "warn", "severity: info|warn|block")
	af.StringVar(&stdCheck, "check", "manual", "built-in check id, or 'manual' for agent-judged")
	registerContent(standardAddCmd)

	standardCmd.AddCommand(standardAddCmd, standardListCmd, standardShowCmd, standardRetireCmd)
	rootCmd.AddCommand(standardCmd)
}
