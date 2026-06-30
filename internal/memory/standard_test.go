package memory

import (
	"strings"
	"testing"
)

func TestAddAndReadStandard(t *testing.T) {
	s := newTestStore(t)
	std, err := s.AddStandard("Target Metric", "Every project declares a metric",
		"tag:backend", "warn", "has-summary", "Summary must name a metric.")
	if err != nil {
		t.Fatalf("AddStandard: %v", err)
	}
	if std.ID != "target-metric" {
		t.Errorf("id = %q, want target-metric", std.ID)
	}

	got, err := s.ReadStandard("target-metric")
	if err != nil {
		t.Fatalf("ReadStandard: %v", err)
	}
	if got.AppliesTo != "tag:backend" || got.Severity != "warn" || got.Check != "has-summary" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.Status != "active" {
		t.Errorf("status = %q, want active", got.Status)
	}
	if !strings.Contains(got.Body, "name a metric") {
		t.Errorf("body lost: %q", got.Body)
	}
}

func TestAddStandardDefaultsAndValidation(t *testing.T) {
	s := newTestStore(t)

	// Defaults: severity=warn, check=manual, applies-to=all.
	std, err := s.AddStandard("plain", "Plain", "", "", "", "")
	if err != nil {
		t.Fatalf("AddStandard defaults: %v", err)
	}
	if std.Severity != "warn" || std.Check != "manual" || std.AppliesTo != "all" {
		t.Errorf("defaults wrong: %+v", std)
	}

	if _, err := s.AddStandard("dup", "Dup", "all", "warn", "manual", ""); err != nil {
		t.Fatalf("first dup add: %v", err)
	}
	if _, err := s.AddStandard("dup", "Dup", "all", "warn", "manual", ""); err == nil {
		t.Error("expected duplicate standard error")
	}
	if _, err := s.AddStandard("badsev", "x", "all", "loud", "manual", ""); err == nil {
		t.Error("expected invalid severity error")
	}
	if _, err := s.AddStandard("badcheck", "x", "all", "warn", "no-such-check", ""); err == nil {
		t.Error("expected invalid check error")
	}
}

func TestRetireStandard(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AddStandard("x", "X", "all", "warn", "manual", ""); err != nil {
		t.Fatal(err)
	}
	std, err := s.RetireStandard("x")
	if err != nil {
		t.Fatalf("RetireStandard: %v", err)
	}
	if std.Status != "retired" {
		t.Errorf("status = %q, want retired", std.Status)
	}
}

func TestAuditStandardsScopeAndManual(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	tagProject(t, s, "beta", "Beta", "backend")
	tagProject(t, s, "gamma", "Gamma") // untagged, out of backend scope
	// alpha gets a real summary; beta keeps the placeholder.
	if _, err := s.Write("alpha", TypeSummary, "A real summary.", WriteOpts{}); err != nil {
		t.Fatal(err)
	}

	// Structural standard scoped to backend.
	if _, err := s.AddStandard("has-summary-std", "Has summary", "tag:backend", "warn", "has-summary", ""); err != nil {
		t.Fatal(err)
	}
	// Manual standard scoped to all.
	if _, err := s.AddStandard("metric", "Metric", "all", "block", "manual", ""); err != nil {
		t.Fatal(err)
	}

	rep, err := s.AuditStandards("", today(t))
	if err != nil {
		t.Fatalf("AuditStandards: %v", err)
	}

	// backend standard → 2 cells (alpha pass, beta fail); manual → 3 cells unknown.
	if rep.Summary[StatusPass] != 1 || rep.Summary[StatusFail] != 1 || rep.Summary[StatusUnknown] != 3 {
		t.Errorf("summary = %v, want pass=1 fail=1 unknown=3", rep.Summary)
	}

	// Retiring the structural standard drops its cells.
	if _, err := s.RetireStandard("has-summary-std"); err != nil {
		t.Fatal(err)
	}
	rep, err = s.AuditStandards("", today(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Matrix) != 3 || rep.Summary[StatusUnknown] != 3 {
		t.Errorf("after retire: matrix=%d summary=%v, want 3 unknown cells", len(rep.Matrix), rep.Summary)
	}
}

func TestAuditStandardsRestrict(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	tagProject(t, s, "beta", "Beta", "backend")
	if _, err := s.AddStandard("std", "Std", "tag:backend", "warn", "has-summary", ""); err != nil {
		t.Fatal(err)
	}

	// Restrict the project axis to a single project.
	rep, err := s.AuditStandard("std", "alpha", today(t))
	if err != nil {
		t.Fatalf("AuditStandard: %v", err)
	}
	if len(rep.Matrix) != 1 || rep.Matrix[0].Project != "alpha" {
		t.Errorf("restrict failed: %+v", rep.Matrix)
	}
	if rep.Matrix[0].Concern != "std" || rep.Matrix[0].Kind != "standard" || rep.Matrix[0].Severity != "warn" {
		t.Errorf("cell metadata wrong: %+v", rep.Matrix[0])
	}
}
