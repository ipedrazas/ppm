package memory

import (
	"os"
	"strings"
)

// Waive records (or overwrites) a reasoned exception of one concern — a standard
// (or, later, an initiative) — for one project. The waiver entry is named by the
// concern id, so there is exactly one waiver per concern per project and audit
// can look it up directly. A waiver makes an intentional exception read as
// "waived" rather than "fail", which is what keeps the matrix free of alert
// fatigue — so a reason is required by the command layer.
func (s *Store) Waive(project, concern, reason string) (*Entry, error) {
	concern = Slugify(concern)
	if concern == "" {
		return nil, memErrf("waiver requires a concern id")
	}
	if fi, err := os.Stat(s.projectDir(project)); err != nil || !fi.IsDir() {
		return nil, memErrf("project %q not found", project)
	}
	return s.Write(project, TypeWaiver, reason, WriteOpts{
		Name:  concern,
		Extra: []KV{{Key: "standard", Val: concern}},
	})
}

// Waivers returns a project's waivers as concern-id → reason. A missing waivers/
// dir yields an empty map.
func (s *Store) Waivers(project string) (map[string]string, error) {
	es, err := s.List(project, TypeWaiver)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(es))
	for _, e := range es {
		concern := fmString(e.Frontmatter, "standard")
		if concern == "" {
			concern = e.Name
		}
		out[concern] = strings.TrimSpace(e.Body)
	}
	return out, nil
}

// waiverFor reports the waiver reason for one concern in one project, if any.
func (s *Store) waiverFor(project, concern string) (string, bool) {
	ws, err := s.Waivers(project)
	if err != nil {
		return "", false
	}
	reason, ok := ws[concern]
	return reason, ok
}
