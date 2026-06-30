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
	auditCheck    string
	auditStandard string
	auditTag      string
	auditProject  string
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Evaluate standards across projects (the compliance matrix)",
	Long: "Emit a cross-project compliance matrix. With no flags, runs every active\n" +
		"standard over its own applies-to scope. Use --standard to run one standard,\n" +
		"or --check to run an ad-hoc built-in check over all projects. Narrow the\n" +
		"project axis with --tag or --project. Built-in checks: has-summary,\n" +
		"has-focus, decisions-link-tasks, active-has-tracker, no-stale-questions:Nd,\n" +
		"freshness:Nd.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		if auditCheck != "" && auditStandard != "" {
			return fmt.Errorf("pass at most one of --check/--standard")
		}
		restrict, err := auditScope()
		if err != nil {
			return err
		}
		now := time.Now().UTC()

		var rep *memory.AuditReport
		switch {
		case auditCheck != "":
			rep, err = st.Audit(auditCheck, restrict, now)
		case auditStandard != "":
			rep, err = st.AuditStandard(auditStandard, restrict, now)
		default:
			rep, err = st.AuditStandards(restrict, now)
		}
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
		return "No matching standards or projects in scope."
	}
	var b strings.Builder
	for _, c := range r.Matrix {
		fmt.Fprintf(&b, "%-7s %s · %s", c.Status, c.Concern, c.Project)
		if c.Severity != "" {
			fmt.Fprintf(&b, " [%s]", c.Severity)
		}
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
	f.StringVar(&auditCheck, "check", "", "run an ad-hoc built-in check, e.g. has-summary | no-stale-questions:14d")
	f.StringVar(&auditStandard, "standard", "", "run a single standard by id (else all active standards)")
	f.StringVar(&auditTag, "tag", "", "restrict the project axis to this tag")
	f.StringVar(&auditProject, "project", "", "restrict the project axis to a single project")
	rootCmd.AddCommand(auditCmd)
}
