package memory

import (
	"os"
	"strings"
)

// ProjectUpdate carries optional edits to a project's index frontmatter. A nil
// field is left unchanged; a non-nil field is set (empty string clears nothing —
// it writes an empty value).
type ProjectUpdate struct {
	Title          *string
	Status         *string
	TrackerSystem  *string
	TrackerProject *string
	TrackerURL     *string
	// AddTags / RemoveTags merge into the index's tags sequence (additive then
	// subtractive), driving tag:<t> scoping for cross-cutting concerns.
	AddTags    []string
	RemoveTags []string
}

// IsEmpty reports whether no field was provided.
func (u ProjectUpdate) IsEmpty() bool {
	return u.Title == nil && u.Status == nil &&
		u.TrackerSystem == nil && u.TrackerProject == nil && u.TrackerURL == nil &&
		len(u.AddTags) == 0 && len(u.RemoveTags) == 0
}

// UpdateProject applies frontmatter edits to a project's index and bumps the
// updated date. Tracker fields are written into a nested tracker mapping.
func (s *Store) UpdateProject(project string, u ProjectUpdate) (*Entry, error) {
	raw, err := s.Read(project, "", "")
	if err != nil {
		return nil, memErrf("project %q not found", project)
	}
	fm, body := ParseDoc(raw)

	if u.Title != nil {
		fm.Set("title", *u.Title)
	}
	if u.Status != nil {
		fm.Set("status", *u.Status)
	}
	if u.TrackerSystem != nil || u.TrackerProject != nil || u.TrackerURL != nil {
		tracker := fm.ensureMap("tracker")
		if u.TrackerSystem != nil {
			setMapScalar(tracker, "system", *u.TrackerSystem)
		}
		if u.TrackerProject != nil {
			setMapScalar(tracker, "project", *u.TrackerProject)
		}
		if u.TrackerURL != nil {
			setMapScalar(tracker, "url", *u.TrackerURL)
		}
	}
	if len(u.AddTags) > 0 || len(u.RemoveTags) > 0 {
		fm.SetSeq("tags", mergeTags(fm.GetSeq("tags"), u.AddTags, u.RemoveTags))
	}
	fm.Set("updated", Today())

	rel := projectRel(project, "index.md")
	if err := os.WriteFile(s.abs(rel), []byte(SerializeDoc(fm, body)), filePerm); err != nil {
		return nil, err
	}
	return &Entry{
		Project:     project,
		Type:        TypeIndex,
		Name:        string(TypeIndex),
		Frontmatter: fm.ToMap(),
		Body:        strings.TrimSpace(body),
		RelPath:     rel,
	}, nil
}

// mergeTags applies an additive-then-subtractive merge over a project's tags:
// existing ∪ add, minus remove, deduplicated, insertion order preserved.
func mergeTags(existing, add, remove []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(existing)+len(add))
	for _, t := range existing {
		if t = strings.TrimSpace(t); t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	for _, t := range add {
		if t = strings.TrimSpace(t); t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	if len(remove) == 0 {
		return out
	}
	rm := map[string]bool{}
	for _, t := range remove {
		rm[strings.TrimSpace(t)] = true
	}
	kept := out[:0]
	for _, t := range out {
		if !rm[t] {
			kept = append(kept, t)
		}
	}
	return kept
}
