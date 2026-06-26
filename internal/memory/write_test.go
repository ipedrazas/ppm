package memory

import (
	"strings"
	"testing"
)

func TestWriteDecisionAutoSlug(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	e, err := s.Write("p", TypeDecision, "Email nudge first; cheap and testable.", WriteOpts{})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.HasPrefix(e.Name, Today()+"-") {
		t.Errorf("name %q missing date prefix", e.Name)
	}
	if e.Frontmatter["type"] != "decision" {
		t.Errorf("type = %v", e.Frontmatter["type"])
	}
	if e.Frontmatter["ts"] == nil || e.Frontmatter["ts"] == "" {
		t.Errorf("missing ts")
	}
}

func TestWriteQuestionDefaultsOpen(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	e, err := s.Write("p", TypeQuestion, "Do analytics exist?", WriteOpts{Name: "analytics"})
	if err != nil {
		t.Fatal(err)
	}
	if e.Frontmatter["status"] != "open" {
		t.Errorf("status = %v, want open", e.Frontmatter["status"])
	}
}

func TestResolveQuestion(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("p", TypeQuestion, "Open?", WriteOpts{Name: "q1"}); err != nil {
		t.Fatal(err)
	}
	e, err := s.Write("p", TypeQuestion, "Resolved answer.", WriteOpts{Name: "q1", Mode: ModeResolve})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if e.Frontmatter["status"] != "resolved" {
		t.Errorf("status = %v, want resolved", e.Frontmatter["status"])
	}
	if e.Body != "Resolved answer." {
		t.Errorf("body = %q", e.Body)
	}
	// created/ts preserved through resolve
	if e.Frontmatter["ts"] == nil {
		t.Errorf("ts lost on resolve")
	}
}

func TestResolveKeepsBodyWhenContentEmpty(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("p", TypeQuestion, "Original body.", WriteOpts{Name: "q"}); err != nil {
		t.Fatal(err)
	}
	e, err := s.Write("p", TypeQuestion, "", WriteOpts{Name: "q", Mode: ModeResolve})
	if err != nil {
		t.Fatal(err)
	}
	if e.Body != "Original body." {
		t.Errorf("body = %q, want preserved original", e.Body)
	}
}

func TestResolveRequiresName(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Write("p", TypeQuestion, "x", WriteOpts{Mode: ModeResolve})
	if !IsMemoryError(err) {
		t.Fatalf("expected MemoryError, got %v", err)
	}
}

func TestWriteTaskExtraFrontmatterOrder(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	e, err := s.Write("p", TypeTask, "Scope: email.", WriteOpts{
		Name:  "ENG-123",
		Extra: []KV{{"ref", "ENG-123"}, {"url", "https://x/ENG-123"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.Frontmatter["ref"] != "ENG-123" {
		t.Errorf("ref = %v", e.Frontmatter["ref"])
	}
	raw, err := s.Read("p", TypeTask, e.Name)
	if err != nil {
		t.Fatal(err)
	}
	refPos := strings.Index(raw, "ref:")
	urlPos := strings.Index(raw, "url:")
	if refPos < 0 || urlPos < 0 || refPos > urlPos {
		t.Errorf("expected ref before url in:\n%s", raw)
	}
}

func TestSingletonReplaces(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("p", TypeSummary, "first", WriteOpts{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("p", TypeSummary, "second", WriteOpts{}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Read("p", TypeSummary, "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "first") || !strings.Contains(got, "second") {
		t.Errorf("summary not replaced:\n%s", got)
	}
}
