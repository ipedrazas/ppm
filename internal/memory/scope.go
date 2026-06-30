package memory

import (
	"slices"
	"strings"
)

// ResolveScope expands an applies-to expression into the matching project slugs,
// in the workspace's sorted project order. Forms:
//
//	""  or  "all"   → every project
//	"tag:<t>"       → projects whose index tags include <t>
//	"a, b, c"       → an explicit, slugified project list (each must exist)
//
// This is the single resolver behind every cross-cutting concern's scope, so a
// standard, an initiative, and an ad-hoc `ppm audit --tag` all agree on
// membership.
func (s *Store) ResolveScope(expr string) ([]string, error) {
	expr = strings.TrimSpace(expr)
	all, err := s.ListProjects()
	if err != nil {
		return nil, err
	}

	switch {
	case expr == "" || expr == "all":
		return all, nil

	case strings.HasPrefix(expr, "tag:"):
		tag := strings.TrimSpace(strings.TrimPrefix(expr, "tag:"))
		out := make([]string, 0, len(all))
		for _, p := range all {
			if s.projectHasTag(p, tag) {
				out = append(out, p)
			}
		}
		return out, nil

	default:
		known := make(map[string]bool, len(all))
		for _, p := range all {
			known[p] = true
		}
		var out []string
		for raw := range strings.SplitSeq(expr, ",") {
			p := Slugify(strings.TrimSpace(raw))
			if p == "" {
				continue
			}
			if !known[p] {
				return nil, memErrf("unknown project %q in scope", p)
			}
			out = append(out, p)
		}
		return out, nil
	}
}

// projectHasTag reports whether a project's index frontmatter tags include tag.
func (s *Store) projectHasTag(project, tag string) bool {
	raw, err := s.Read(project, "", "")
	if err != nil {
		return false
	}
	fm, _ := ParseDoc(raw)
	return slices.Contains(fm.GetSeq("tags"), tag)
}
