package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextAssembly(t *testing.T) {
	s := newTestStore(t)
	// workspace globals
	if err := os.WriteFile(filepath.Join(s.Root(), "preferences.md"), []byte("# preferences\nask when unclear"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := s.CreateProject("active", "Active project"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateProject("other", "Other project"); err != nil {
		t.Fatal(err)
	}

	_, _ = s.Write("active", TypeSummary, "the stable overview", WriteOpts{})
	_, _ = s.Write("active", TypeFocus, "current thread", WriteOpts{})
	for _, n := range []string{"d1", "d2", "d3", "d4"} {
		_, _ = s.Write("active", TypeDecision, "body "+n, WriteOpts{Name: n})
	}
	_, _ = s.Write("active", TypeQuestion, "still open", WriteOpts{Name: "q-open"})
	_, _ = s.Write("active", TypeQuestion, "to close", WriteOpts{Name: "q-done"})
	_, _ = s.Write("active", TypeQuestion, "closed", WriteOpts{Name: "q-done", Mode: ModeResolve})

	ctx, err := s.Context("active", 2)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ctx.Preferences, "ask when unclear") {
		t.Errorf("preferences not loaded: %q", ctx.Preferences)
	}
	// summary/focus are stripped to prose (no frontmatter)
	if strings.Contains(ctx.Summary, "type:") || ctx.Summary != "the stable overview" {
		t.Errorf("summary = %q, want bare prose", ctx.Summary)
	}
	if ctx.Focus != "current thread" {
		t.Errorf("focus = %q", ctx.Focus)
	}
	// index retains frontmatter
	if !strings.Contains(ctx.Index, "title: Active project") {
		t.Errorf("index missing frontmatter: %q", ctx.Index)
	}
	// only open questions, full entries
	if len(ctx.OpenQuestions) != 1 || ctx.OpenQuestions[0].Name != "q-open" {
		t.Errorf("openQuestions = %v", names(ctx.OpenQuestions))
	}
	// recent decisions capped at 2, newest first
	if len(ctx.RecentDecisions) != 2 || ctx.RecentDecisions[0].Name != "d4" {
		t.Errorf("recentDecisions = %v, want [d4 d3]", names(ctx.RecentDecisions))
	}
	// shape still counts all 4 decisions
	if ctx.Shape.Counts[TypeDecision] != 4 {
		t.Errorf("shape decision count = %d, want 4", ctx.Shape.Counts[TypeDecision])
	}
	// other project appears as a one-liner
	if len(ctx.OtherProjects) != 1 || ctx.OtherProjects[0].Project != "other" {
		t.Errorf("otherProjects = %v", ctx.OtherProjects)
	}
	if ctx.OtherProjects[0].Title != "Other project" {
		t.Errorf("other title = %q", ctx.OtherProjects[0].Title)
	}
}

func TestContextMissingProject(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Context("ghost", 3); !IsMemoryError(err) {
		t.Fatalf("expected MemoryError, got %v", err)
	}
}
