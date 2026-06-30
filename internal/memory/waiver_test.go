package memory

import "testing"

func TestWaiveRoundTrip(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("billing", "Billing"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Waive("billing", "gdpr-2026", "B2B only, no PII."); err != nil {
		t.Fatalf("Waive: %v", err)
	}
	ws, err := s.Waivers("billing")
	if err != nil {
		t.Fatalf("Waivers: %v", err)
	}
	if ws["gdpr-2026"] != "B2B only, no PII." {
		t.Errorf("waiver reason = %q", ws["gdpr-2026"])
	}
}

func TestWaiveUnknownProject(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Waive("nope", "std", "x"); err == nil {
		t.Fatal("expected error waiving into a nonexistent project")
	}
}

func TestWaiveOverwrites(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("billing", "Billing"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Waive("billing", "std", "first reason"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Waive("billing", "std", "updated reason"); err != nil {
		t.Fatal(err)
	}
	ws, _ := s.Waivers("billing")
	if len(ws) != 1 || ws["std"] != "updated reason" {
		t.Errorf("re-waive should overwrite: %v", ws)
	}
}

func TestAuditWaiverOverlay(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	tagProject(t, s, "beta", "Beta", "backend")
	// Both keep the placeholder summary → both fail has-summary.
	if _, err := s.AddStandard("std", "Has summary", "tag:backend", "warn", "has-summary", ""); err != nil {
		t.Fatal(err)
	}

	// Baseline: two fails.
	rep, err := s.AuditStandard("std", "", today(t))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Summary[StatusFail] != 2 {
		t.Fatalf("baseline summary = %v, want fail=2", rep.Summary)
	}

	// Waive beta → one fail, one waived.
	if _, err := s.Waive("beta", "std", "legacy service, summary tracked elsewhere"); err != nil {
		t.Fatal(err)
	}
	rep, err = s.AuditStandard("std", "", today(t))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Summary[StatusFail] != 1 || rep.Summary[StatusWaived] != 1 {
		t.Errorf("after waive: summary = %v, want fail=1 waived=1", rep.Summary)
	}
	for _, c := range rep.Matrix {
		if c.Project == "beta" {
			if c.Status != StatusWaived || c.Reason == "" {
				t.Errorf("beta cell = %+v, want waived with a reason", c)
			}
		}
	}
}

func TestAuditWaiverDoesNotMaskPass(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	if _, err := s.Write("alpha", TypeSummary, "A real summary.", WriteOpts{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddStandard("std", "Has summary", "tag:backend", "warn", "has-summary", ""); err != nil {
		t.Fatal(err)
	}
	// A waiver on a passing project is moot — the cell stays pass.
	if _, err := s.Waive("alpha", "std", "redundant"); err != nil {
		t.Fatal(err)
	}
	rep, err := s.AuditStandard("std", "", today(t))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Summary[StatusPass] != 1 || rep.Summary[StatusWaived] != 0 {
		t.Errorf("summary = %v, want pass=1 waived=0", rep.Summary)
	}
}
