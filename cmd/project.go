package cmd

import (
	"fmt"
	"strings"

	"github.com/ipedrazas/ppm/internal/memory"
	"github.com/ipedrazas/ppm/internal/output"
	"github.com/spf13/cobra"
)

var projectTitle string

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Create and inspect projects",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <slug>",
	Short: "Create a project, scaffolding its index/summary/focus",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		slug, err := st.CreateProject(args[0], projectTitle)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Created project \"" + slug + "\".",
			Data:    map[string]string{"project": slug, "title": projectTitle},
		})
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		projects, err := st.ListProjects()
		if err != nil {
			return err
		}
		msg := "No projects yet."
		if len(projects) > 0 {
			msg = "Projects:"
			for _, p := range projects {
				msg += "\n- " + p
			}
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: msg,
			Data:    map[string]any{"projects": projects},
		})
	},
}

var projectShowCmd = &cobra.Command{
	Use:   "show <slug>",
	Short: "Show a project's shape (entry inventory, no content)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		shape, err := st.Shape(args[0])
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: formatShape(shape),
			Data:    shape,
		})
	},
}

func formatShape(s *memory.ProjectShape) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%s)\n", s.Title, s.Status)
	b.WriteString("counts:")
	if len(s.Counts) == 0 {
		b.WriteString(" (none)")
	}
	for _, t := range memory.CollectionTypes {
		if n := s.Counts[t]; n > 0 {
			fmt.Fprintf(&b, " %s=%d", t, n)
		}
	}
	for _, e := range s.Entries {
		b.WriteByte('\n')
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

var (
	updTitle          string
	updStatus         string
	updTrackerSystem  string
	updTrackerProject string
	updTrackerURL     string
)

var projectUpdateCmd = &cobra.Command{
	Use:   "update <slug>",
	Short: "Edit a project's index frontmatter (status/title/tracker)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := openStore()
		if err != nil {
			return err
		}
		up := memory.ProjectUpdate{}
		f := cmd.Flags()
		if f.Changed("title") {
			up.Title = &updTitle
		}
		if f.Changed("status") {
			up.Status = &updStatus
		}
		if f.Changed("tracker-system") {
			up.TrackerSystem = &updTrackerSystem
		}
		if f.Changed("tracker-project") {
			up.TrackerProject = &updTrackerProject
		}
		if f.Changed("tracker-url") {
			up.TrackerURL = &updTrackerURL
		}
		if up.IsEmpty() {
			return fmt.Errorf("nothing to update: pass at least one of --title/--status/--tracker-*")
		}
		entry, err := st.UpdateProject(args[0], up)
		if err != nil {
			return err
		}
		return emit(cmd, output.Result{
			OK:      true,
			Message: "Updated project \"" + args[0] + "\".",
			Data:    entry,
		})
	},
}

func init() {
	projectCreateCmd.Flags().StringVar(&projectTitle, "title", "", "human-readable project title (required)")
	_ = projectCreateCmd.MarkFlagRequired("title")

	uf := projectUpdateCmd.Flags()
	uf.StringVar(&updTitle, "title", "", "project title")
	uf.StringVar(&updStatus, "status", "", "status: active|paused|done|archived")
	uf.StringVar(&updTrackerSystem, "tracker-system", "", "tracker system, e.g. linear|jira")
	uf.StringVar(&updTrackerProject, "tracker-project", "", "tracker project name")
	uf.StringVar(&updTrackerURL, "tracker-url", "", "tracker project URL")

	projectCmd.AddCommand(projectCreateCmd, projectListCmd, projectShowCmd, projectUpdateCmd)
	rootCmd.AddCommand(projectCmd)
}
