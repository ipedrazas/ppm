package memory

import (
	"os"
	"path"
	"sort"
	"strings"
)

// Standard is one declarative cross-cutting invariant: a built-in check (or
// "manual") bound to an applies-to scope and a severity. Standards live in the
// workspace standards/ collection, parallel to projects/.
type Standard struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	AppliesTo string `json:"appliesTo"`
	Severity  string `json:"severity"`
	Check     string `json:"check"`
	Status    string `json:"status"`
	Body      string `json:"body"`
	RelPath   string `json:"relPath"`
}

// validSeverities is the closed set for a standard's severity.
var validSeverities = map[string]bool{"info": true, "warn": true, "block": true}

func (s *Store) standardsDir() string {
	return s.abs(WorkspaceRegistries[TypeStandard].At)
}

func standardRel(id string) string {
	return path.Join(WorkspaceRegistries[TypeStandard].At, id+".md")
}

// AddStandard creates a standard, erroring if one with the id already exists.
// Empty severity/check/applies-to default to warn/manual/all. A non-manual check
// must be a known built-in; an unknown severity or check is rejected.
func (s *Store) AddStandard(id, title, appliesTo, severity, check, body string) (*Standard, error) {
	id = Slugify(id)
	if id == "" {
		return nil, memErrf("standard id is required")
	}
	if severity == "" {
		severity = "warn"
	}
	if !validSeverities[severity] {
		return nil, memErrf("invalid severity %q: want info|warn|block", severity)
	}
	if check == "" {
		check = "manual"
	}
	if check != "manual" {
		if _, err := ResolveCheck(check); err != nil {
			return nil, err
		}
	}
	if appliesTo == "" {
		appliesTo = "all"
	}

	if err := os.MkdirAll(s.standardsDir(), dirPerm); err != nil {
		return nil, err
	}
	rel := standardRel(id)
	if _, err := os.Stat(s.abs(rel)); err == nil {
		return nil, memErrf("standard %q already exists", id)
	}

	fm := NewFrontmatter()
	fm.Set("type", string(TypeStandard))
	fm.Set("id", id)
	fm.Set("title", title)
	fm.Set("applies-to", appliesTo)
	fm.Set("severity", severity)
	fm.Set("check", check)
	fm.Set("status", "active")
	fm.Set("created", Today())
	fm.Set("updated", Today())
	if err := os.WriteFile(s.abs(rel), []byte(SerializeDoc(fm, body)), filePerm); err != nil {
		return nil, err
	}
	return &Standard{
		ID: id, Title: title, AppliesTo: appliesTo, Severity: severity,
		Check: check, Status: "active", Body: strings.TrimSpace(body), RelPath: rel,
	}, nil
}

// ReadStandard loads one standard by id.
func (s *Store) ReadStandard(id string) (*Standard, error) {
	rel := standardRel(Slugify(id))
	raw, err := os.ReadFile(s.abs(rel))
	if err != nil {
		return nil, memErrf("standard %q not found", id)
	}
	return parseStandard(raw, rel), nil
}

// ListStandards returns all standards sorted by id. A missing dir yields empty.
func (s *Store) ListStandards() ([]Standard, error) {
	ents, err := os.ReadDir(s.standardsDir())
	if err != nil {
		return []Standard{}, nil
	}
	out := make([]Standard, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		rel := path.Join(WorkspaceRegistries[TypeStandard].At, e.Name())
		raw, err := os.ReadFile(s.abs(rel))
		if err != nil {
			return nil, err
		}
		out = append(out, *parseStandard(raw, rel))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// RetireStandard flips a standard's status to retired (kept for history, never
// deleted) and bumps its updated date. Retired standards are skipped by audit.
func (s *Store) RetireStandard(id string) (*Standard, error) {
	rel := standardRel(Slugify(id))
	raw, err := os.ReadFile(s.abs(rel))
	if err != nil {
		return nil, memErrf("standard %q not found", id)
	}
	fm, body := ParseDoc(string(raw))
	fm.Set("status", "retired")
	fm.Set("updated", Today())
	doc := SerializeDoc(fm, body)
	if err := os.WriteFile(s.abs(rel), []byte(doc), filePerm); err != nil {
		return nil, err
	}
	return parseStandard([]byte(doc), rel), nil
}

func parseStandard(raw []byte, rel string) *Standard {
	fm, body := ParseDoc(string(raw))
	return &Standard{
		ID:        orElse(fm, "id", strings.TrimSuffix(path.Base(rel), ".md")),
		Title:     orElse(fm, "title", ""),
		AppliesTo: orElse(fm, "applies-to", "all"),
		Severity:  orElse(fm, "severity", "warn"),
		Check:     orElse(fm, "check", "manual"),
		Status:    orElse(fm, "status", "active"),
		Body:      body,
		RelPath:   rel,
	}
}
