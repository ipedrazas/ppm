package memory

import (
	"os"
	"time"
)

// ProjectLine is a one-line reference to a project, so cross-references resolve
// without injecting the project's content.
type ProjectLine struct {
	Project string `json:"project"`
	Title   string `json:"title"`
	Status  string `json:"status"`
}

// Context is the shape-aware slice injected each turn: full content for the
// cheap high-value entries of the active project, shape only for the rest, and
// one-liners for every other project. This mirrors the spec's transformContext.
type Context struct {
	Project         string        `json:"project"`
	Preferences     string        `json:"preferences"`
	Glossary        string        `json:"glossary"`
	Index           string        `json:"index"`
	Summary         string        `json:"summary"`
	Focus           string        `json:"focus"`
	OpenQuestions   []Entry       `json:"openQuestions"`
	RecentDecisions []Entry       `json:"recentDecisions"`
	Shape           *ProjectShape `json:"shape"`
	OtherProjects   []ProjectLine `json:"otherProjects"`
	// Standards and Initiatives are the cross-cutting concerns whose scope
	// includes this project, each with its current audit status — so the agent
	// sees its obligations every turn, not only when audit is run by hand.
	Standards   []AuditCell `json:"standards"`
	Initiatives []AuditCell `json:"initiatives"`
}

// ReadGlobal returns a workspace file's content (preferences/glossary/index),
// or "" if it does not exist.
func (s *Store) ReadGlobal(name string) string {
	b, err := os.ReadFile(s.abs(name + ".md"))
	if err != nil {
		return ""
	}
	return string(b)
}

// Context assembles the injected slice for the active project. recentDecisions
// caps how many decisions are included with full content. It errors if the
// project does not exist.
func (s *Store) Context(project string, recentDecisions int) (*Context, error) {
	if fi, err := os.Stat(s.projectDir(project)); err != nil || !fi.IsDir() {
		return nil, memErrf("project %q not found", project)
	}

	ctx := &Context{
		Project:     project,
		Preferences: s.ReadGlobal("preferences"),
		Glossary:    s.ReadGlobal("glossary"),
	}
	// Index keeps its frontmatter (id/title/status/tracker is valuable context);
	// summary/focus are plain prose, so strip their (type-only) frontmatter.
	if v, err := s.Read(project, "", ""); err == nil {
		ctx.Index = v
	}
	if v, err := s.Read(project, TypeSummary, ""); err == nil {
		_, ctx.Summary = ParseDoc(v)
	}
	if v, err := s.Read(project, TypeFocus, ""); err == nil {
		_, ctx.Focus = ParseDoc(v)
	}

	oq, err := s.OpenQuestions(project)
	if err != nil {
		return nil, err
	}
	ctx.OpenQuestions = oq

	rd, err := s.Recent(project, TypeDecision, recentDecisions)
	if err != nil {
		return nil, err
	}
	ctx.RecentDecisions = rd

	shape, err := s.Shape(project)
	if err != nil {
		return nil, err
	}
	ctx.Shape = shape

	projects, _ := s.ListProjects()
	for _, p := range projects {
		if p == project {
			continue
		}
		title, status := s.projectMeta(p)
		ctx.OtherProjects = append(ctx.OtherProjects, ProjectLine{
			Project: p,
			Title:   title,
			Status:  status,
		})
	}

	// Cross-cutting obligations on this project: every active standard and
	// initiative whose scope includes it, with its current status.
	if rep, err := s.AuditAll(project, time.Now().UTC()); err == nil {
		for _, c := range rep.Matrix {
			if c.Kind == "initiative" {
				ctx.Initiatives = append(ctx.Initiatives, c)
			} else {
				ctx.Standards = append(ctx.Standards, c)
			}
		}
	}
	return ctx, nil
}

// projectMeta reads a project's title/status from its index frontmatter.
func (s *Store) projectMeta(project string) (title, status string) {
	raw, err := s.Read(project, "", "")
	if err != nil {
		return project, ""
	}
	fm, _ := ParseDoc(raw)
	return orElse(fm, "title", project), orElse(fm, "status", "active")
}
