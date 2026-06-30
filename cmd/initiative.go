package cmd

import (
	"fmt"
	"strings"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var (
	initTitle     string
	initAppliesTo string
	initStatus    string
	initRef       string
	initURL       string
)

var initiativeCmd = &cobra.Command{
	Use:     "initiative",
	Aliases: []string{"init"},
	Short:   "Manage cross-project initiatives (campaigns spanning projects)",
}

var initiativeAddCmd = &cobra.Command{
	Use:   "add <id>",
	Short: "Declare an initiative: a campaign over a set of projects",
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
		init, err := st.AddInitiative(args[0], initTitle, initAppliesTo, body)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Declared initiative \"" + init.ID + "\".",
			Data:    init,
		})
	},
}

var initiativeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all initiatives",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		inits, err := st.ListInitiatives()
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatInitiatives(inits),
			Data:    map[string]any{"initiatives": inits},
		})
	},
}

var initiativeShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show an initiative with its per-member bound/unbound rollup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		roll, err := st.Rollup(args[0])
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatRollup(roll),
			Data:    roll,
		})
	},
}

var initiativeBindCmd = &cobra.Command{
	Use:   "bind <id> <project>",
	Short: "Bind a project to an initiative by scaffolding a backlinked task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		body, err := contentInput.resolveOptional()
		if err != nil {
			return err
		}
		entry, err := st.BindInitiative(args[0], args[1], initRef, initURL, body)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: fmt.Sprintf("Bound %q to initiative %q → %s", args[1], args[0], entry.RelPath),
			Data:    entry,
		})
	},
}

var initiativeUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an initiative's status (active|paused|done)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		if initStatus == "" {
			return fmt.Errorf("nothing to update: pass --status active|paused|done")
		}
		init, err := st.SetInitiativeStatus(args[0], initStatus)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Updated initiative \"" + init.ID + "\".",
			Data:    init,
		})
	},
}

func formatInitiatives(inits []memory.Initiative) string {
	if len(inits) == 0 {
		return "No initiatives defined."
	}
	var b strings.Builder
	b.WriteString("Initiatives:")
	for _, i := range inits {
		b.WriteByte('\n')
		fmt.Fprintf(&b, "- %s [%s] → %s", i.ID, i.Status, i.AppliesTo)
		if i.Title != "" {
			fmt.Fprintf(&b, ": %s", i.Title)
		}
	}
	return b.String()
}

func formatRollup(r *memory.InitiativeRollup) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%s) → %s\n", r.ID, r.Status, r.AppliesTo)
	fmt.Fprintf(&b, "bound %d/%d members", r.BoundCount, len(r.Members))
	for _, m := range r.Members {
		b.WriteByte('\n')
		if m.Bound {
			fmt.Fprintf(&b, "- ✓ %s → %s", m.Project, m.Task)
		} else {
			fmt.Fprintf(&b, "- ✗ %s (unbound)", m.Project)
		}
	}
	return b.String()
}

func init() {
	af := initiativeAddCmd.Flags()
	af.StringVar(&initTitle, "title", "", "human-readable initiative title")
	af.StringVar(&initAppliesTo, "applies-to", "all", "scope: all | tag:<t> | comma-separated slugs")
	registerContent(initiativeAddCmd)

	bf := initiativeBindCmd.Flags()
	bf.StringVar(&initRef, "ref", "", "tracker reference for the member task, e.g. ENG-411 (required)")
	bf.StringVar(&initURL, "url", "", "tracker URL")
	_ = initiativeBindCmd.MarkFlagRequired("ref")
	registerContent(initiativeBindCmd)

	initiativeUpdateCmd.Flags().StringVar(&initStatus, "status", "", "status: active|paused|done")

	initiativeCmd.AddCommand(initiativeAddCmd, initiativeListCmd, initiativeShowCmd,
		initiativeBindCmd, initiativeUpdateCmd)
	rootCmd.AddCommand(initiativeCmd)
}
