package memory

import (
	"reflect"
	"testing"
)

// tagProject creates a project and tags it.
func tagProject(t *testing.T, s *Store, slug, title string, tags ...string) {
	t.Helper()
	if _, err := s.CreateProject(slug, title); err != nil {
		t.Fatalf("CreateProject %s: %v", slug, err)
	}
	if len(tags) > 0 {
		if _, err := s.UpdateProject(Slugify(slug), ProjectUpdate{AddTags: tags}); err != nil {
			t.Fatalf("tag %s: %v", slug, err)
		}
	}
}

func TestResolveScope(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "backend")
	tagProject(t, s, "beta", "Beta", "backend", "customer-facing")
	tagProject(t, s, "gamma", "Gamma")

	cases := []struct {
		name string
		expr string
		want []string
	}{
		{"all", "all", []string{"alpha", "beta", "gamma"}},
		{"empty means all", "", []string{"alpha", "beta", "gamma"}},
		{"tag backend", "tag:backend", []string{"alpha", "beta"}},
		{"tag customer-facing", "tag:customer-facing", []string{"beta"}},
		{"tag none match", "tag:missing", []string{}},
		{"explicit list", "alpha, gamma", []string{"alpha", "gamma"}},
		{"explicit slugifies", "Alpha", []string{"alpha"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := s.ResolveScope(c.expr)
			if err != nil {
				t.Fatalf("ResolveScope(%q): %v", c.expr, err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("ResolveScope(%q) = %v, want %v", c.expr, got, c.want)
			}
		})
	}
}

func TestResolveScopeUnknownProject(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha")
	if _, err := s.ResolveScope("alpha, nope"); err == nil {
		t.Fatal("expected error for unknown project in explicit scope")
	}
}

func TestTagMergeAddThenRemove(t *testing.T) {
	s := newTestStore(t)
	tagProject(t, s, "alpha", "Alpha", "a", "b")
	// Adding a dup and a new tag, removing one.
	if _, err := s.UpdateProject("alpha", ProjectUpdate{AddTags: []string{"b", "c"}, RemoveTags: []string{"a"}}); err != nil {
		t.Fatalf("update: %v", err)
	}
	raw, _ := s.Read("alpha", "", "")
	fm, _ := ParseDoc(raw)
	got := fm.GetSeq("tags")
	want := []string{"b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tags = %v, want %v", got, want)
	}
}
