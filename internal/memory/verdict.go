package memory

import (
	"os"
	"strings"
)

// Verdict is a recorded judgement of a manual standard for one project: pass or
// fail, plus rationale. It lets a manual standard resolve beyond "unknown".
type Verdict struct {
	Standard string `json:"standard"`
	Status   string `json:"status"`
	Reason   string `json:"reason"`
}

// validVerdictStatus is the closed set for a recorded judgement.
var validVerdictStatus = map[string]bool{"pass": true, "fail": true}

// RecordVerdict writes (or overwrites) the agent's judgement of a manual standard
// for one project. Named by the standard id, so there is one verdict per standard
// per project and audit can look it up directly.
func (s *Store) RecordVerdict(project, standard, status, reason string) (*Entry, error) {
	standard = Slugify(standard)
	if standard == "" {
		return nil, memErrf("verdict requires a standard id")
	}
	if !validVerdictStatus[status] {
		return nil, memErrf("invalid verdict %q: want pass|fail", status)
	}
	if fi, err := os.Stat(s.projectDir(project)); err != nil || !fi.IsDir() {
		return nil, memErrf("project %q not found", project)
	}
	return s.Write(project, TypeVerdict, reason, WriteOpts{
		Name: standard,
		Extra: []KV{
			{Key: "standard", Val: standard},
			{Key: "verdict", Val: status},
		},
	})
}

// verdictFor returns the recorded judgement for one standard in one project.
func (s *Store) verdictFor(project, standard string) (Verdict, bool) {
	es, err := s.List(project, TypeVerdict)
	if err != nil {
		return Verdict{}, false
	}
	for _, e := range es {
		std := fmString(e.Frontmatter, "standard")
		if std == "" {
			std = e.Name
		}
		if std != standard {
			continue
		}
		return Verdict{
			Standard: std,
			Status:   fmString(e.Frontmatter, "verdict"),
			Reason:   strings.TrimSpace(e.Body),
		}, true
	}
	return Verdict{}, false
}
