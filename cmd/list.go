package cmd

import (
	"fmt"
	"strings"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var (
	listRecent   int
	listOpenOnly bool
)

// runList lists a collection type's entries, honoring --recent and (for
// questions) --open.
func runList(cmd *cobra.Command, project string, t memory.EntryType) error {
	st, err := openStore()
	if err != nil {
		return err
	}

	var entries []memory.Entry
	switch {
	case t == memory.TypeQuestion && listOpenOnly:
		entries, err = st.OpenQuestions(project)
	case listRecent > 0:
		entries, err = st.Recent(project, t, listRecent)
	default:
		entries, err = st.List(project, t)
	}
	if err != nil {
		return err
	}

	var b strings.Builder
	if len(entries) == 0 {
		fmt.Fprintf(&b, "No %ss.", t)
	}
	for i, e := range entries {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(formatEntryLine(e))
	}
	return emit(cmd, output.Result{
		OK:      true,
		Message: b.String(),
		Data:    map[string]any{"entries": entries},
	})
}

// formatEntryLine renders "- name (date) [status]: title".
func formatEntryLine(e memory.Entry) string {
	var b strings.Builder
	b.WriteString("- ")
	b.WriteString(e.Name)
	if d, ok := e.Frontmatter["created"].(string); ok && d != "" {
		fmt.Fprintf(&b, " (%s)", d)
	}
	if s, ok := e.Frontmatter["status"].(string); ok && s != "" {
		fmt.Fprintf(&b, " [%s]", s)
	}
	title := memory.Title(e.Body)
	if title == "" {
		title = e.Name
	}
	b.WriteString(": ")
	b.WriteString(title)
	return b.String()
}

// addListCommand attaches a "list <project>" subcommand for a collection type.
func addListCommand(parent *cobra.Command, t memory.EntryType, withOpen bool) {
	c := &cobra.Command{
		Use:   "list <project>",
		Short: fmt.Sprintf("List %s entries", t),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, args[0], t)
		},
	}
	c.Flags().IntVar(&listRecent, "recent", 0, "show only the N most recent")
	if withOpen {
		c.Flags().BoolVar(&listOpenOnly, "open", false, "show only open questions")
	}
	parent.AddCommand(c)
}
