package memory

import (
	"path/filepath"
	"strings"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s := NewStore(filepath.Join(t.TempDir(), "memory"))
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return s
}

func TestInitScaffolds(t *testing.T) {
	s := newTestStore(t)
	for _, f := range []string{"index.md", "preferences.md", "glossary.md"} {
		if _, err := readFile(filepath.Join(s.Root(), f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}
}

func TestCreateProjectRoundTrip(t *testing.T) {
	s := newTestStore(t)
	slug, err := s.CreateProject("Onboarding Drop-off", "Onboarding drop-off")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if slug != "onboarding-drop-off" {
		t.Fatalf("slug = %q", slug)
	}

	idx, err := s.Read(slug, "", "")
	if err != nil {
		t.Fatalf("Read index: %v", err)
	}
	fm, _ := ParseDoc(idx)
	if got, _ := fm.Get("title"); got != "Onboarding drop-off" {
		t.Errorf("title = %q", got)
	}
	if got, _ := fm.Get("status"); got != "active" {
		t.Errorf("status = %q", got)
	}

	summary, err := s.Read(slug, TypeSummary, "")
	if err != nil {
		t.Fatalf("Read summary: %v", err)
	}
	if !strings.Contains(summary, "type: summary") {
		t.Errorf("summary missing type frontmatter:\n%s", summary)
	}
}

func TestCreateProjectDuplicate(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.CreateProject("dup", "Dup"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := s.CreateProject("dup", "Dup")
	if err == nil || !IsMemoryError(err) {
		t.Fatalf("expected MemoryError on duplicate, got %v", err)
	}
}

func TestReadWorkspaceIndex(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Read("", "", ""); err != nil {
		t.Fatalf("Read workspace index: %v", err)
	}
}

func TestWriteUnknownType(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Write("p", EntryType("bogus"), "x", WriteOpts{}); !IsMemoryError(err) {
		t.Fatalf("expected MemoryError, got %v", err)
	}
}

func TestListProjectsEmpty(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "memory"))
	got, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}
