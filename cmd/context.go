package cmd

import (
	"fmt"
	"strings"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var contextRecent int

var contextCmd = &cobra.Command{
	Use:   "context <project>",
	Short: "Emit the shape-aware context slice injected each turn",
	Long: "Assembles the injected slice: workspace preferences/glossary, the active\n" +
		"project's index/summary/focus, all open questions and the N most recent\n" +
		"decisions (full content), the rest of the project as shape only, and every\n" +
		"other project as a one-liner.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		ctx, err := st.Context(args[0], contextRecent)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatContext(ctx),
			Data:    ctx,
		})
	},
}

func formatContext(c *memory.Context) string {
	var b strings.Builder
	section := func(title, body string) {
		body = strings.TrimSpace(body)
		if body == "" {
			return
		}
		fmt.Fprintf(&b, "# %s\n%s\n\n", title, body)
	}

	section("preferences", c.Preferences)
	section("glossary", c.Glossary)

	fmt.Fprintf(&b, "## Active project: %s (%s)\n", c.Shape.Title, c.Shape.Status)
	if idx := strings.TrimSpace(c.Index); idx != "" {
		b.WriteString(idx + "\n\n")
	}
	section("summary", c.Summary)
	section("focus", c.Focus)

	fmt.Fprintf(&b, "### open questions (%d)\n", len(c.OpenQuestions))
	for _, e := range c.OpenQuestions {
		fmt.Fprintf(&b, "- %s: %s\n", e.Name, entryTitle(e))
	}
	b.WriteByte('\n')

	total := c.Shape.Counts[memory.TypeDecision]
	fmt.Fprintf(&b, "### recent decisions (%d shown of %d)\n", len(c.RecentDecisions), total)
	for _, e := range c.RecentDecisions {
		fmt.Fprintf(&b, "--- %s\n%s\n", e.Name, strings.TrimSpace(e.Body))
	}
	b.WriteByte('\n')

	b.WriteString("### shape\n")
	b.WriteString(shapeInventory(c.Shape))

	if len(c.Standards) > 0 || len(c.Initiatives) > 0 {
		b.WriteString("\n\n### cross-cutting obligations")
		for _, c := range c.Standards {
			fmt.Fprintf(&b, "\n- standard %s [%s]: %s", c.Concern, c.Severity, c.Status)
			if c.Reason != "" {
				fmt.Fprintf(&b, " — %s", c.Reason)
			}
		}
		for _, c := range c.Initiatives {
			fmt.Fprintf(&b, "\n- initiative %s: %s", c.Concern, c.Status)
			if c.Reason != "" {
				fmt.Fprintf(&b, " — %s", c.Reason)
			}
			if c.Detail != "" {
				fmt.Fprintf(&b, " (%s)", c.Detail)
			}
		}
	}

	if len(c.OtherProjects) > 0 {
		b.WriteString("\n\n## Other projects\n")
		for _, p := range c.OtherProjects {
			fmt.Fprintf(&b, "- %s: %s (%s)\n", p.Project, p.Title, p.Status)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func entryTitle(e memory.Entry) string {
	if t := memory.Title(e.Body); t != "" {
		return t
	}
	return e.Name
}

// shapeInventory renders the per-entry inventory lines of a shape.
func shapeInventory(s *memory.ProjectShape) string {
	var b strings.Builder
	for i, e := range s.Entries {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "- %s/%s", e.Type, e.Name)
		if e.Date != "" {
			fmt.Fprintf(&b, " (%s)", e.Date)
		}
		if e.Status != "" {
			fmt.Fprintf(&b, " [%s]", e.Status)
		}
		fmt.Fprintf(&b, ": %s", e.Title)
	}
	return b.String()
}

func init() {
	contextCmd.Flags().IntVar(&contextRecent, "recent", 3, "decisions to include with full content")
	rootCmd.AddCommand(contextCmd)
}
