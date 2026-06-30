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
	auditCheck      string
	auditStandard   string
	auditInitiative string
	auditTag        string
	auditProject    string
	auditStrict     bool
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Evaluate standards across projects (the compliance matrix)",
	Long: "Emit a cross-project compliance matrix. With no flags, runs every active\n" +
		"standard and initiative over its own applies-to scope. Use --standard or\n" +
		"--initiative to run one, or --check for an ad-hoc built-in check over all\n" +
		"projects. Narrow the project axis with --tag or --project. Built-in checks:\n" +
		"has-summary, has-focus, decisions-link-tasks, active-has-tracker,\n" +
		"no-stale-questions:Nd, freshness:Nd.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		if err := exactlyOneAuditTarget(); err != nil {
			return err
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
		case auditInitiative != "":
			rep, err = st.AuditInitiative(auditInitiative, restrict)
		default:
			rep, err = st.AuditAll(restrict, now)
		}
		if err != nil {
			return err
		}
		// --strict: render the matrix as usual, but exit non-zero if any cell
		// failed — for CI gating. Waived/unknown do not trip it.
		if auditStrict && rep.Summary[memory.StatusFail] > 0 {
			pendingExit = 1
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatAudit(rep),
			Data:    rep,
		})
	},
}

// exactlyOneAuditTarget rejects combining the mutually exclusive target flags
// (--check/--standard/--initiative); zero is fine and means "all".
func exactlyOneAuditTarget() error {
	n := 0
	for _, v := range []string{auditCheck, auditStandard, auditInitiative} {
		if v != "" {
			n++
		}
	}
	if n > 1 {
		return fmt.Errorf("pass at most one of --check/--standard/--initiative")
	}
	return nil
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
	wConcern := 0
	for _, c := range r.Matrix {
		if len(c.Concern) > wConcern {
			wConcern = len(c.Concern)
		}
	}
	var b strings.Builder
	for _, c := range r.Matrix {
		fmt.Fprintf(&b, "%-7s %-*s  %s", c.Status, wConcern, c.Concern, c.Project)
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
	f.StringVar(&auditStandard, "standard", "", "run a single standard by id")
	f.StringVar(&auditInitiative, "initiative", "", "run a single initiative by id")
	f.StringVar(&auditTag, "tag", "", "restrict the project axis to this tag")
	f.StringVar(&auditProject, "project", "", "restrict the project axis to a single project")
	f.BoolVar(&auditStrict, "strict", false, "exit non-zero if any cell failed (for CI gating)")
	rootCmd.AddCommand(auditCmd)
}
