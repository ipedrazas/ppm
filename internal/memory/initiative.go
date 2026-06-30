package memory

import (
	"os"
	"path"
	"sort"
	"strings"
)

// Initiative is a workspace-level cross-project campaign: a scope plus a status.
// It holds the why and the membership rollup; the actual per-project work is a
// normal task entry that backlinks to the initiative (so live status stays in the
// tracker, never here).
type Initiative struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	AppliesTo string `json:"appliesTo"`
	Status    string `json:"status"`
	Body      string `json:"body"`
	RelPath   string `json:"relPath"`
}

// validInitiativeStatus is the closed set for an initiative's lifecycle.
var validInitiativeStatus = map[string]bool{"active": true, "paused": true, "done": true}

func (s *Store) initiativesDir() string {
	return s.abs(WorkspaceRegistries[TypeInitiative].At)
}

func initiativeRel(id string) string {
	return path.Join(WorkspaceRegistries[TypeInitiative].At, id+".md")
}

// initiativeLink is the wikilink a member task carries to bind it to an
// initiative; audit detects membership by scanning task bodies for it.
func initiativeLink(id string) string { return "[[initiatives/" + id + "]]" }

// AddInitiative creates an initiative, erroring if the id already exists. Empty
// applies-to defaults to all.
func (s *Store) AddInitiative(id, title, appliesTo, body string) (*Initiative, error) {
	id = Slugify(id)
	if id == "" {
		return nil, memErrf("initiative id is required")
	}
	if appliesTo == "" {
		appliesTo = "all"
	}
	if err := os.MkdirAll(s.initiativesDir(), dirPerm); err != nil {
		return nil, err
	}
	rel := initiativeRel(id)
	if _, err := os.Stat(s.abs(rel)); err == nil {
		return nil, memErrf("initiative %q already exists", id)
	}

	fm := NewFrontmatter()
	fm.Set("type", string(TypeInitiative))
	fm.Set("id", id)
	fm.Set("title", title)
	fm.Set("applies-to", appliesTo)
	fm.Set("status", "active")
	fm.Set("created", Today())
	fm.Set("updated", Today())
	if err := os.WriteFile(s.abs(rel), []byte(SerializeDoc(fm, body)), filePerm); err != nil {
		return nil, err
	}
	return &Initiative{
		ID: id, Title: title, AppliesTo: appliesTo, Status: "active",
		Body: strings.TrimSpace(body), RelPath: rel,
	}, nil
}

// ReadInitiative loads one initiative by id.
func (s *Store) ReadInitiative(id string) (*Initiative, error) {
	rel := initiativeRel(Slugify(id))
	raw, err := os.ReadFile(s.abs(rel))
	if err != nil {
		return nil, memErrf("initiative %q not found", id)
	}
	return parseInitiative(raw, rel), nil
}

// ListInitiatives returns all initiatives sorted by id.
func (s *Store) ListInitiatives() ([]Initiative, error) {
	ents, err := os.ReadDir(s.initiativesDir())
	if err != nil {
		return []Initiative{}, nil
	}
	out := make([]Initiative, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		rel := path.Join(WorkspaceRegistries[TypeInitiative].At, e.Name())
		raw, err := os.ReadFile(s.abs(rel))
		if err != nil {
			return nil, err
		}
		out = append(out, *parseInitiative(raw, rel))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// SetInitiativeStatus updates an initiative's lifecycle status and bumps updated.
func (s *Store) SetInitiativeStatus(id, status string) (*Initiative, error) {
	if !validInitiativeStatus[status] {
		return nil, memErrf("invalid status %q: want active|paused|done", status)
	}
	rel := initiativeRel(Slugify(id))
	raw, err := os.ReadFile(s.abs(rel))
	if err != nil {
		return nil, memErrf("initiative %q not found", id)
	}
	fm, body := ParseDoc(string(raw))
	fm.Set("status", status)
	fm.Set("updated", Today())
	doc := SerializeDoc(fm, body)
	if err := os.WriteFile(s.abs(rel), []byte(doc), filePerm); err != nil {
		return nil, err
	}
	return parseInitiative([]byte(doc), rel), nil
}

// BindInitiative scaffolds a member task in project, referencing the tracker
// (ref/url) and backlinking to the initiative so audit counts it as bound. The
// task slug defaults to the ref. Re-binding the same ref overwrites it.
func (s *Store) BindInitiative(id, project, ref, url, content string) (*Entry, error) {
	init, err := s.ReadInitiative(id)
	if err != nil {
		return nil, err
	}
	if ref == "" {
		return nil, memErrf("bind requires a tracker --ref")
	}
	if fi, err := os.Stat(s.projectDir(project)); err != nil || !fi.IsDir() {
		return nil, memErrf("project %q not found", project)
	}

	body := strings.TrimSpace(content)
	link := initiativeLink(init.ID)
	if !strings.Contains(body, link) {
		if body != "" {
			body += "\n\n"
		}
		body += "→ " + link
	}

	extra := []KV{{Key: "ref", Val: ref}}
	if url != "" {
		extra = append(extra, KV{Key: "url", Val: url})
	}
	return s.Write(project, TypeTask, body, WriteOpts{Name: ref, Extra: extra})
}

// InitiativeMember is one project's membership status in an initiative.
type InitiativeMember struct {
	Project string `json:"project"`
	Bound   bool   `json:"bound"`
	Task    string `json:"task,omitempty"`
}

// InitiativeRollup is an initiative plus its per-member bound/unbound status.
type InitiativeRollup struct {
	Initiative
	Members    []InitiativeMember `json:"members"`
	BoundCount int                `json:"boundCount"`
}

// Rollup resolves an initiative's scope and reports, per member project, whether
// a task backlinks to it.
func (s *Store) Rollup(id string) (*InitiativeRollup, error) {
	init, err := s.ReadInitiative(id)
	if err != nil {
		return nil, err
	}
	projects, err := s.ResolveScope(init.AppliesTo)
	if err != nil {
		return nil, err
	}
	out := &InitiativeRollup{Initiative: *init}
	for _, p := range projects {
		task, bound := s.boundTask(p, init.ID)
		out.Members = append(out.Members, InitiativeMember{Project: p, Bound: bound, Task: task})
		if bound {
			out.BoundCount++
		}
	}
	return out, nil
}

// boundTask returns the name of the first task in project that backlinks to the
// initiative, if any.
func (s *Store) boundTask(project, initiativeID string) (string, bool) {
	tasks, err := s.List(project, TypeTask)
	if err != nil {
		return "", false
	}
	link := initiativeLink(initiativeID)
	for _, t := range tasks {
		if strings.Contains(t.Body, link) {
			return t.Name, true
		}
	}
	return "", false
}

func parseInitiative(raw []byte, rel string) *Initiative {
	fm, body := ParseDoc(string(raw))
	return &Initiative{
		ID:        orElse(fm, "id", strings.TrimSuffix(path.Base(rel), ".md")),
		Title:     orElse(fm, "title", ""),
		AppliesTo: orElse(fm, "applies-to", "all"),
		Status:    orElse(fm, "status", "active"),
		Body:      body,
		RelPath:   rel,
	}
}
