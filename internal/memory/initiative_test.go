package memory

import (
	"strings"
	"testing"
)

func TestAddAndReadInitiative(t *testing.T) {
	s := newTestStore(t)
	init, err := s.AddInitiative("GDPR 2026", "GDPR review", "tag:customer-facing", "Each member needs a review.")
	if err != nil {
		t.Fatalf("AddInitiative: %v", err)
	}
	if init.ID != "gdpr-2026" {
		t.Errorf("id = %q, want gdpr-2026", init.ID)
	}
	got, err := s.ReadInitiative("gdpr-2026")
	if err != nil {
		t.Fatalf("ReadInitiative: %v", err)
	}
	if got.AppliesTo != "tag:customer-facing" || got.Status != "active" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestInitiativeStatusValidation(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddInitiative("x", "X", "all", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetInitiativeStatus("x", "done"); err != nil {
		t.Fatalf("SetInitiativeStatus done: %v", err)
	}
	if _, err := s.SetInitiativeStatus("x", "bogus"); err == nil {
		t.Error("expected invalid status error")
	}
	if _, err := s.AddInitiative("x", "X", "all", ""); err == nil {
		t.Error("expected duplicate initiative error")
	}
}

func TestBindInitiativeAndRollup(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "customer-facing")
	tagProject(t, s, "beta", "Beta", "customer-facing")
	if _, err := s.AddInitiative("gdpr", "GDPR", "tag:customer-facing", ""); err != nil {
		t.Fatal(err)
	}

	// Bind alpha only.
	entry, err := s.BindInitiative("gdpr", "alpha", "ENG-411", "https://x/411", "Data review.")
	if err != nil {
		t.Fatalf("BindInitiative: %v", err)
	}
	if !strings.Contains(entry.Body, "[[initiatives/gdpr]]") {
		t.Errorf("bound task missing backlink: %q", entry.Body)
	}
	if entry.Name != "eng-411" { // Write slugifies entry names
		t.Errorf("task name = %q, want eng-411", entry.Name)
	}

	roll, err := s.Rollup("gdpr")
	if err != nil {
		t.Fatalf("Rollup: %v", err)
	}
	if roll.BoundCount != 1 || len(roll.Members) != 2 {
		t.Errorf("rollup = bound %d / %d members, want 1/2", roll.BoundCount, len(roll.Members))
	}
	for _, m := range roll.Members {
		switch m.Project {
		case "alpha":
			if !m.Bound || m.Task != "eng-411" {
				t.Errorf("alpha member = %+v, want bound to eng-411", m)
			}
		case "beta":
			if m.Bound {
				t.Errorf("beta should be unbound: %+v", m)
			}
		}
	}
}

func TestBindRequiresRef(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha")
	if _, err := s.AddInitiative("x", "X", "all", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BindInitiative("x", "alpha", "", "", "body"); err == nil {
		t.Error("expected error binding without a ref")
	}
}

func TestAuditInitiatives(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "customer-facing")
	tagProject(t, s, "beta", "Beta", "customer-facing")
	if _, err := s.AddInitiative("gdpr", "GDPR", "tag:customer-facing", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BindInitiative("gdpr", "alpha", "ENG-1", "", ""); err != nil {
		t.Fatal(err)
	}

	rep, err := s.AuditInitiatives("", today(t))
	if err != nil {
		t.Fatalf("AuditInitiatives: %v", err)
	}
	if rep.Summary[StatusPass] != 1 || rep.Summary[StatusFail] != 1 {
		t.Errorf("summary = %v, want pass=1 fail=1", rep.Summary)
	}

	// Waiving the unbound member flips its fail to waived.
	if _, err := s.Waive("beta", "gdpr", "B2B only, no PII"); err != nil {
		t.Fatal(err)
	}
	rep, err = s.AuditInitiatives("", today(t))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Summary[StatusPass] != 1 || rep.Summary[StatusWaived] != 1 || rep.Summary[StatusFail] != 0 {
		t.Errorf("after waive: summary = %v, want pass=1 waived=1 fail=0", rep.Summary)
	}

	// A paused initiative is skipped by the all-initiatives audit.
	if _, err := s.SetInitiativeStatus("gdpr", "paused"); err != nil {
		t.Fatal(err)
	}
	rep, _ = s.AuditInitiatives("", today(t))
	if len(rep.Matrix) != 0 {
		t.Errorf("paused initiative should be skipped, got %d cells", len(rep.Matrix))
	}
}

func TestVerdictResolvesManualStandard(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	if _, err := s.AddStandard("metric", "Names a metric", "tag:backend", "block", "manual", ""); err != nil {
		t.Fatal(err)
	}

	// Before a verdict: unknown.
	rep, err := s.AuditStandard("metric", "", today(t))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Summary[StatusUnknown] != 1 {
		t.Fatalf("before verdict: summary = %v, want unknown=1", rep.Summary)
	}

	// Record a pass verdict → cell becomes pass.
	if _, err := s.RecordVerdict("alpha", "metric", "pass", "Summary names DAU target."); err != nil {
		t.Fatalf("RecordVerdict: %v", err)
	}
	rep, err = s.AuditStandard("metric", "", today(t))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Summary[StatusPass] != 1 || rep.Summary[StatusUnknown] != 0 {
		t.Errorf("after verdict: summary = %v, want pass=1 unknown=0", rep.Summary)
	}
	if got := rep.Matrix[0].Reason; got != "Summary names DAU target." {
		t.Errorf("verdict reason = %q", got)
	}
}

func TestVerdictValidation(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha")
	if _, err := s.RecordVerdict("alpha", "std", "maybe", "x"); err == nil {
		t.Error("expected invalid verdict status error")
	}
	if _, err := s.RecordVerdict("nope", "std", "pass", "x"); err == nil {
		t.Error("expected unknown-project error")
	}
}

func TestAuditAllCombinesStandardsAndInitiatives(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	if _, err := s.Write("alpha", TypeSummary, "Real summary.", WriteOpts{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddStandard("has-summary", "Has summary", "tag:backend", "warn", "has-summary", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddInitiative("gdpr", "GDPR", "tag:backend", ""); err != nil {
		t.Fatal(err)
	}

	rep, err := s.AuditAll("", today(t))
	if err != nil {
		t.Fatalf("AuditAll: %v", err)
	}
	// standard passes (real summary), initiative fails (no bound task).
	if rep.Summary[StatusPass] != 1 || rep.Summary[StatusFail] != 1 {
		t.Errorf("summary = %v, want pass=1 fail=1", rep.Summary)
	}
	var kinds []string
	for _, c := range rep.Matrix {
		kinds = append(kinds, c.Kind)
	}
	if len(kinds) != 2 {
		t.Errorf("expected 2 cells (standard+initiative), got %v", kinds)
	}
}
