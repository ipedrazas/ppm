package memory

import (
	"strconv"
	"strings"
	"time"
)

// AuditStatus is one cell's verdict in a cross-project audit.
type AuditStatus string

const (
	StatusPass    AuditStatus = "pass"
	StatusFail    AuditStatus = "fail"
	StatusWaived  AuditStatus = "waived"
	StatusUnknown AuditStatus = "unknown"
	StatusNA      AuditStatus = "n/a"
)

// Placeholder bodies written by CreateProject; a project that still carries them
// has not filled in that singleton, so the has-summary / has-focus checks fail.
const (
	placeholderSummary = "_To be written._"
	placeholderFocus   = "_No current focus yet._"
)

// AuditCell is one (check, project) result in the compliance matrix.
type AuditCell struct {
	Check   string      `json:"check"`
	Project string      `json:"project"`
	Status  AuditStatus `json:"status"`
	Reason  string      `json:"reason,omitempty"`
	Detail  string      `json:"detail,omitempty"`
}

// AuditReport is the compliance matrix plus a status rollup.
type AuditReport struct {
	Matrix  []AuditCell         `json:"matrix"`
	Summary map[AuditStatus]int `json:"summary"`
}

// checkResult is the outcome of one built-in check on one project.
type checkResult struct {
	status AuditStatus
	reason string
	detail string
}

func pass() checkResult                      { return checkResult{status: StatusPass} }
func fail(reason, detail string) checkResult { return checkResult{StatusFail, reason, detail} }

// builtinCheck evaluates one project; now anchors the date-relative checks.
type builtinCheck func(s *Store, project string, now time.Time) checkResult

// ResolveCheck parses a check id — a name with an optional ":param" suffix, e.g.
// "no-stale-questions:14d" — into an evaluator. Unknown names are an error so a
// typo never silently passes everything.
func ResolveCheck(id string) (builtinCheck, error) {
	name, param, _ := strings.Cut(id, ":")
	switch name {
	case "has-summary":
		return checkHasSummary, nil
	case "has-focus":
		return checkHasFocus, nil
	case "decisions-link-tasks":
		return checkDecisionsLinkTasks, nil
	case "active-has-tracker":
		return checkActiveHasTracker, nil
	case "no-stale-questions":
		days, err := parseDays(param, 14)
		if err != nil {
			return nil, err
		}
		return func(s *Store, p string, now time.Time) checkResult {
			return checkNoStaleQuestions(s, p, now, days)
		}, nil
	case "freshness":
		days, err := parseDays(param, 30)
		if err != nil {
			return nil, err
		}
		return func(s *Store, p string, now time.Time) checkResult {
			return checkFreshness(s, p, now, days)
		}, nil
	default:
		return nil, memErrf("unknown check %q", id)
	}
}

// parseDays reads an "Nd" or "N" day count, defaulting when the suffix is absent.
func parseDays(param string, def int) (int, error) {
	param = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(param), "d"))
	if param == "" {
		return def, nil
	}
	n, err := strconv.Atoi(param)
	if err != nil || n < 0 {
		return 0, memErrf("invalid day count %q", param)
	}
	return n, nil
}

// Audit runs one built-in check across the projects matching scope and returns
// the compliance matrix with a status rollup. Pass time.Now().UTC() for now.
func (s *Store) Audit(checkID, scope string, now time.Time) (*AuditReport, error) {
	check, err := ResolveCheck(checkID)
	if err != nil {
		return nil, err
	}
	projects, err := s.ResolveScope(scope)
	if err != nil {
		return nil, err
	}
	rep := &AuditReport{Summary: map[AuditStatus]int{}}
	for _, p := range projects {
		r := check(s, p, now)
		rep.Matrix = append(rep.Matrix, AuditCell{
			Check:   checkID,
			Project: p,
			Status:  r.status,
			Reason:  r.reason,
			Detail:  r.detail,
		})
		rep.Summary[r.status]++
	}
	return rep, nil
}

func checkHasSummary(s *Store, project string, _ time.Time) checkResult {
	raw, err := s.Read(project, TypeSummary, "")
	if err != nil {
		return fail("no summary file", "")
	}
	if _, body := ParseDoc(raw); body == "" || body == placeholderSummary {
		return fail("summary is a placeholder", "")
	}
	return pass()
}

func checkHasFocus(s *Store, project string, _ time.Time) checkResult {
	raw, err := s.Read(project, TypeFocus, "")
	if err != nil {
		return fail("no focus file", "")
	}
	if _, body := ParseDoc(raw); body == "" || body == placeholderFocus {
		return fail("focus is a placeholder", "")
	}
	return pass()
}

func checkDecisionsLinkTasks(s *Store, project string, _ time.Time) checkResult {
	ds, err := s.List(project, TypeDecision)
	if err != nil {
		return fail("cannot list decisions", "")
	}
	var missing []string
	for _, d := range ds {
		if !strings.Contains(d.Body, "[[tasks/") {
			missing = append(missing, d.Name)
		}
	}
	if len(missing) > 0 {
		return fail("decisions without a task link", strings.Join(missing, ", "))
	}
	return pass()
}

func checkActiveHasTracker(s *Store, project string, _ time.Time) checkResult {
	raw, err := s.Read(project, "", "")
	if err != nil {
		return fail("no index", "")
	}
	fm, _ := ParseDoc(raw)
	if orElse(fm, "status", "active") != "active" {
		return checkResult{status: StatusNA, reason: "not active"}
	}
	if !fm.HasMapping("tracker") {
		return fail("active project has no tracker", "")
	}
	return pass()
}

func checkNoStaleQuestions(s *Store, project string, now time.Time, days int) checkResult {
	oq, err := s.OpenQuestions(project)
	if err != nil {
		return fail("cannot list questions", "")
	}
	cutoff := now.AddDate(0, 0, -days)
	var stale []string
	for _, q := range oq {
		created := fmString(q.Frontmatter, "created")
		t, err := time.Parse("2006-01-02", created)
		if err != nil {
			continue // undated questions can't be judged stale
		}
		if t.Before(cutoff) {
			stale = append(stale, q.Name)
		}
	}
	if len(stale) > 0 {
		return fail("open questions older than "+strconv.Itoa(days)+"d", strings.Join(stale, ", "))
	}
	return pass()
}

func checkFreshness(s *Store, project string, now time.Time, days int) checkResult {
	raw, err := s.Read(project, "", "")
	if err != nil {
		return fail("no index", "")
	}
	fm, _ := ParseDoc(raw)
	updated, _ := fm.Get("updated")
	t, err := time.Parse("2006-01-02", updated)
	if err != nil {
		return fail("no valid updated date", "")
	}
	if t.Before(now.AddDate(0, 0, -days)) {
		return fail("index not updated in "+strconv.Itoa(days)+"d", "updated "+updated)
	}
	return pass()
}
