package memory

import (
	"strings"
	"testing"
)

func strptr(s string) *string { return &s }

func TestUpdateProjectStatusAndTracker(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	e, err := s.UpdateProject("p", ProjectUpdate{
		Status:         strptr("paused"),
		TrackerSystem:  strptr("linear"),
		TrackerProject: strptr("Onboarding"),
		TrackerURL:     strptr("https://linear.app/acme/x"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.Frontmatter["status"] != "paused" {
		t.Errorf("status = %v", e.Frontmatter["status"])
	}
	tr, ok := e.Frontmatter["tracker"].(map[string]any)
	if !ok || tr["system"] != "linear" || tr["project"] != "Onboarding" {
		t.Fatalf("tracker not set correctly: %#v", e.Frontmatter["tracker"])
	}

	// Re-read from disk: nested tracker and original keys must persist.
	raw, err := s.Read("p", "", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"id: p", "status: paused", "tracker:", "system: linear", "url: https://linear.app/acme/x"} {
		if !strings.Contains(raw, want) {
			t.Errorf("index missing %q:\n%s", want, raw)
		}
	}
}

func TestUpdateProjectPartialPreservesOthers(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "Original Title"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateProject("p", ProjectUpdate{Status: strptr("done")}); err != nil {
		t.Fatal(err)
	}
	raw, _ := s.Read("p", "", "")
	if !strings.Contains(raw, "title: Original Title") {
		t.Errorf("title not preserved:\n%s", raw)
	}
	if !strings.Contains(raw, "status: done") {
		t.Errorf("status not updated:\n%s", raw)
	}
}

func TestUpdateProjectMissing(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.UpdateProject("ghost", ProjectUpdate{Status: strptr("done")}); !IsMemoryError(err) {
		t.Fatalf("expected MemoryError, got %v", err)
	}
}
