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

// AuditCell is one (concern, project) result in the compliance matrix. Concern
// is the standard id (Kind "standard") or, for an ad-hoc `--check`, the check id
// (Kind "check"). Check holds the resolved built-in check; Severity is carried
// from the standard.
type AuditCell struct {
	Concern  string      `json:"concern"`
	Kind     string      `json:"kind"`
	Check    string      `json:"check,omitempty"`
	Severity string      `json:"severity,omitempty"`
	Project  string      `json:"project"`
	Status   AuditStatus `json:"status"`
	Reason   string      `json:"reason,omitempty"`
	Detail   string      `json:"detail,omitempty"`
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

// Audit runs one ad-hoc built-in check across the projects matching scope (the
// `ppm audit --check` path) and returns the compliance matrix with a status
// rollup. Pass time.Now().UTC() for now.
func (s *Store) Audit(checkID, scope string, now time.Time) (*AuditReport, error) {
	check, err := ResolveCheck(checkID)
	if err != nil {
		return nil, err
	}
	projects, err := s.ResolveScope(scope)
	if err != nil {
		return nil, err
	}
	rep := newReport()
	for _, p := range projects {
		r := check(s, p, now)
		rep.add(AuditCell{
			Concern: checkID,
			Kind:    "check",
			Check:   checkID,
			Project: p,
			Status:  r.status,
			Reason:  r.reason,
			Detail:  r.detail,
		})
	}
	return rep, nil
}

// AuditStandards runs every active standard over its own applies-to scope,
// optionally intersected with restrict (from --tag/--project), and returns the
// combined matrix. A "manual" standard yields unknown cells for the agent to
// judge. Retired standards are skipped.
func (s *Store) AuditStandards(restrict string, now time.Time) (*AuditReport, error) {
	stds, err := s.ListStandards()
	if err != nil {
		return nil, err
	}
	rep := newReport()
	for _, std := range stds {
		if std.Status == "retired" {
			continue
		}
		cells, err := s.evalStandard(std, restrict, now)
		if err != nil {
			return nil, err
		}
		for _, c := range cells {
			rep.add(c)
		}
	}
	return rep, nil
}

// AuditStandard runs a single standard by id over its scope (intersected with
// restrict), even if it is retired.
func (s *Store) AuditStandard(id, restrict string, now time.Time) (*AuditReport, error) {
	std, err := s.ReadStandard(id)
	if err != nil {
		return nil, err
	}
	cells, err := s.evalStandard(*std, restrict, now)
	if err != nil {
		return nil, err
	}
	rep := newReport()
	for _, c := range cells {
		rep.add(c)
	}
	return rep, nil
}

// evalStandard resolves a standard's scope (intersected with restrict) and
// evaluates its check per project. A manual or unresolvable check yields unknown.
func (s *Store) evalStandard(std Standard, restrict string, now time.Time) ([]AuditCell, error) {
	projects, err := s.ResolveScope(std.AppliesTo)
	if err != nil {
		return nil, err
	}
	if restrict != "" && restrict != "all" {
		allowed, err := s.ResolveScope(restrict)
		if err != nil {
			return nil, err
		}
		projects = intersect(projects, allowed)
	}

	manual := std.Check == "" || std.Check == "manual"
	var check builtinCheck
	if !manual {
		if check, err = ResolveCheck(std.Check); err != nil {
			manual = true // a stored invalid check shouldn't nuke the whole audit
		}
	}

	cells := make([]AuditCell, 0, len(projects))
	for _, p := range projects {
		cell := AuditCell{
			Concern:  std.ID,
			Kind:     "standard",
			Check:    std.Check,
			Severity: std.Severity,
			Project:  p,
		}
		if manual {
			cell.Status, cell.Reason = StatusUnknown, "manual check"
		} else {
			r := check(s, p, now)
			cell.Status, cell.Reason, cell.Detail = r.status, r.reason, r.detail
		}
		cells = append(cells, cell)
	}
	return cells, nil
}

func newReport() *AuditReport {
	return &AuditReport{Summary: map[AuditStatus]int{}}
}

func (r *AuditReport) add(c AuditCell) {
	r.Matrix = append(r.Matrix, c)
	r.Summary[c.Status]++
}

func intersect(a, b []string) []string {
	set := make(map[string]bool, len(b))
	for _, x := range b {
		set[x] = true
	}
	var out []string
	for _, x := range a {
		if set[x] {
			out = append(out, x)
		}
	}
	return out
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
