package memory

import (
	"testing"
	"time"
)

// auditCell runs one check over one project and returns its single cell.
func auditCell(t *testing.T, s *Store, check, project string, now time.Time) AuditCell {
	t.Helper()
	rep, err := s.Audit(check, project, now)
	if err != nil {
		t.Fatalf("Audit(%q, %q): %v", check, project, err)
	}
	if len(rep.Matrix) != 1 {
		t.Fatalf("Audit(%q, %q): got %d cells, want 1", check, project, len(rep.Matrix))
	}
	return rep.Matrix[0]
}

func today(t *testing.T) time.Time {
	t.Helper()
	now, err := time.Parse("2006-01-02", Today())
	if err != nil {
		t.Fatalf("parse today: %v", err)
	}
	return now
}

func TestCheckHasSummary(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("alpha", "Alpha"); err != nil {
		t.Fatal(err)
	}
	now := today(t)

	if got := auditCell(t, s, "has-summary", "alpha", now); got.Status != StatusFail {
		t.Errorf("placeholder summary: status = %q, want fail", got.Status)
	}
	if _, err := s.Write("alpha", TypeSummary, "Reduce drop-off via nudges.", WriteOpts{}); err != nil {
		t.Fatal(err)
	}
	if got := auditCell(t, s, "has-summary", "alpha", now); got.Status != StatusPass {
		t.Errorf("filled summary: status = %q, want pass", got.Status)
	}
}

func TestCheckDecisionsLinkTasks(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("alpha", "Alpha"); err != nil {
		t.Fatal(err)
	}
	now := today(t)

	if _, err := s.Write("alpha", TypeDecision, "Email nudge first.", WriteOpts{Name: "nudge"}); err != nil {
		t.Fatal(err)
	}
	if got := auditCell(t, s, "decisions-link-tasks", "alpha", now); got.Status != StatusFail {
		t.Errorf("unlinked decision: status = %q, want fail", got.Status)
	}
	if _, err := s.Write("alpha", TypeDecision, "Defer dashboard. → [[tasks/ENG-1]]", WriteOpts{Name: "defer"}); err != nil {
		t.Fatal(err)
	}
	// Still fails: the first decision is unlinked.
	if got := auditCell(t, s, "decisions-link-tasks", "alpha", now); got.Status != StatusFail {
		t.Errorf("one unlinked decision remains: status = %q, want fail", got.Status)
	}
}

func TestCheckActiveHasTracker(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("alpha", "Alpha"); err != nil {
		t.Fatal(err)
	}
	now := today(t)

	if got := auditCell(t, s, "active-has-tracker", "alpha", now); got.Status != StatusFail {
		t.Errorf("active, no tracker: status = %q, want fail", got.Status)
	}
	sys := "linear"
	if _, err := s.UpdateProject("alpha", ProjectUpdate{TrackerSystem: &sys}); err != nil {
		t.Fatal(err)
	}
	if got := auditCell(t, s, "active-has-tracker", "alpha", now); got.Status != StatusPass {
		t.Errorf("active, with tracker: status = %q, want pass", got.Status)
	}
	paused := "paused"
	if _, err := s.UpdateProject("alpha", ProjectUpdate{Status: &paused}); err != nil {
		t.Fatal(err)
	}
	if got := auditCell(t, s, "active-has-tracker", "alpha", now); got.Status != StatusNA {
		t.Errorf("paused project: status = %q, want n/a", got.Status)
	}
}

func TestCheckNoStaleQuestions(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("alpha", "Alpha"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("alpha", TypeQuestion, "Do funnel analytics exist?", WriteOpts{Name: "funnel"}); err != nil {
		t.Fatal(err)
	}
	base := today(t)

	// Same day: not stale.
	if got := auditCell(t, s, "no-stale-questions:14d", "alpha", base); got.Status != StatusPass {
		t.Errorf("fresh question: status = %q, want pass", got.Status)
	}
	// 30 days later with a 14d window: stale.
	later := base.AddDate(0, 0, 30)
	if got := auditCell(t, s, "no-stale-questions:14d", "alpha", later); got.Status != StatusFail {
		t.Errorf("aged question: status = %q, want fail", got.Status)
	}
}

func TestCheckFreshness(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("alpha", "Alpha"); err != nil {
		t.Fatal(err)
	}
	base := today(t)

	if got := auditCell(t, s, "freshness:30d", "alpha", base); got.Status != StatusPass {
		t.Errorf("just-updated index: status = %q, want pass", got.Status)
	}
	if got := auditCell(t, s, "freshness:30d", "alpha", base.AddDate(0, 0, 60)); got.Status != StatusFail {
		t.Errorf("stale index: status = %q, want fail", got.Status)
	}
}

func TestAuditMatrixAndSummary(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	tagProject(t, s, "beta", "Beta", "backend")
	// alpha gets a real summary; beta keeps the placeholder.
	if _, err := s.Write("alpha", TypeSummary, "A real summary.", WriteOpts{}); err != nil {
		t.Fatal(err)
	}

	rep, err := s.Audit("has-summary", "tag:backend", today(t))
	if err != nil {
		t.Fatalf("Audit: %v", err)
	}
	if len(rep.Matrix) != 2 {
		t.Fatalf("matrix len = %d, want 2", len(rep.Matrix))
	}
	if rep.Summary[StatusPass] != 1 || rep.Summary[StatusFail] != 1 {
		t.Errorf("summary = %v, want pass=1 fail=1", rep.Summary)
	}
}

func TestResolveCheckUnknown(t *testing.T) {
	if _, err := ResolveCheck("no-such-check"); err == nil {
		t.Fatal("expected error for unknown check")
	}
}

func TestParseDays(t *testing.T) {
	cases := []struct {
		in      string
		def     int
		want    int
		wantErr bool
	}{
		{"", 14, 14, false},
		{"7d", 14, 7, false},
		{"7", 14, 7, false},
		{"bad", 14, 0, true},
		{"-3", 14, 0, true},
	}
	for _, c := range cases {
		got, err := parseDays(c.in, c.def)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseDays(%q): want error", c.in)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Errorf("parseDays(%q) = %d, %v; want %d", c.in, got, err, c.want)
		}
	}
}
