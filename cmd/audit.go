package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var (
	auditCheck   string
	auditTag     string
	auditProject string
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Evaluate a built-in check across projects (the compliance matrix)",
	Long: "Run a built-in structural check over every in-scope project and emit a\n" +
		"compliance matrix. Scope defaults to all projects; narrow it with --tag or\n" +
		"--project. Checks: has-summary, has-focus, decisions-link-tasks,\n" +
		"active-has-tracker, no-stale-questions:Nd, freshness:Nd.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		scope, err := auditScope()
		if err != nil {
			return err
		}
		rep, err := st.Audit(auditCheck, scope, time.Now().UTC())
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatAudit(rep),
			Data:    rep,
		})
	},
}

// auditScope reconciles --tag/--project into a single applies-to expression.
func auditScope() (string, error) {
	if auditTag != "" && auditProject != "" {
		return "", fmt.Errorf("pass at most one of --tag/--project")
	}
	switch {
	case auditTag != "":
		return "tag:" + auditTag, nil
	case auditProject != "":
		return auditProject, nil
	default:
		return "all", nil
	}
}

// auditStatusOrder is the stable display order for the summary rollup.
var auditStatusOrder = []memory.AuditStatus{
	memory.StatusPass, memory.StatusFail, memory.StatusWaived,
	memory.StatusUnknown, memory.StatusNA,
}

func formatAudit(r *memory.AuditReport) string {
	if len(r.Matrix) == 0 {
		return "No projects in scope."
	}
	var b strings.Builder
	for _, c := range r.Matrix {
		fmt.Fprintf(&b, "%-7s %s", c.Status, c.Project)
		if c.Reason != "" {
			fmt.Fprintf(&b, " — %s", c.Reason)
		}
		if c.Detail != "" {
			fmt.Fprintf(&b, " (%s)", c.Detail)
		}
		b.WriteByte('\n')
	}
	var parts []string
	for _, st := range auditStatusOrder {
		if n := r.Summary[st]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", st, n))
		}
	}
	b.WriteString(strings.Join(parts, " "))
	return b.String()
}

func init() {
	f := auditCmd.Flags()
	f.StringVar(&auditCheck, "check", "", "built-in check id, e.g. has-summary | no-stale-questions:14d (required)")
	f.StringVar(&auditTag, "tag", "", "scope to projects carrying this tag")
	f.StringVar(&auditProject, "project", "", "scope to a single project")
	_ = auditCmd.MarkFlagRequired("check")
	rootCmd.AddCommand(auditCmd)
}
