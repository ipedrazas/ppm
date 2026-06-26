package memory

import (
	"testing"
)

func seedProject(t *testing.T) *Store {
	t.Helper()
	s := newTestStore(t)
	if _, err := s.CreateProject("p", "P"); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestListNewestFirst(t *testing.T) {
	s := seedProject(t)
	for _, n := range []string{"d1", "d2", "d3"} {
		if _, err := s.Write("p", TypeDecision, "body "+n, WriteOpts{Name: n}); err != nil {
			t.Fatal(err)
		}
	}
	es, err := s.List("p", TypeDecision)
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 3 {
		t.Fatalf("got %d entries", len(es))
	}
	if es[0].Name != "d3" || es[2].Name != "d1" {
		t.Errorf("order = %s,%s,%s; want d3,d2,d1", es[0].Name, es[1].Name, es[2].Name)
	}
}

func TestRecentLimit(t *testing.T) {
	s := seedProject(t)
	for _, n := range []string{"a", "b", "c", "d"} {
		if _, err := s.Write("p", TypeDecision, "x", WriteOpts{Name: n}); err != nil {
			t.Fatal(err)
		}
	}
	es, err := s.Recent("p", TypeDecision, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 2 {
		t.Fatalf("got %d, want 2", len(es))
	}
	if es[0].Name != "d" || es[1].Name != "c" {
		t.Errorf("recent = %s,%s; want d,c", es[0].Name, es[1].Name)
	}
}

func TestOpenQuestions(t *testing.T) {
	s := seedProject(t)
	if _, err := s.Write("p", TypeQuestion, "open one", WriteOpts{Name: "q1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("p", TypeQuestion, "to resolve", WriteOpts{Name: "q2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Write("p", TypeQuestion, "done", WriteOpts{Name: "q2", Mode: ModeResolve}); err != nil {
		t.Fatal(err)
	}
	open, err := s.OpenQuestions("p")
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 1 || open[0].Name != "q1" {
		t.Errorf("open = %v, want [q1]", names(open))
	}
}

func TestShapeCounts(t *testing.T) {
	s := seedProject(t)
	_, _ = s.Write("p", TypeDecision, "d", WriteOpts{Name: "d1"})
	_, _ = s.Write("p", TypeQuestion, "q", WriteOpts{Name: "q1"})
	_, _ = s.Write("p", TypeTask, "t", WriteOpts{Name: "t1", Extra: []KV{{"ref", "T-1"}}})

	shape, err := s.Shape("p")
	if err != nil {
		t.Fatal(err)
	}
	if shape.Title != "P" || shape.Status != "active" {
		t.Errorf("title/status = %q/%q", shape.Title, shape.Status)
	}
	if shape.Counts[TypeDecision] != 1 || shape.Counts[TypeQuestion] != 1 || shape.Counts[TypeTask] != 1 {
		t.Errorf("counts = %v", shape.Counts)
	}
	if shape.Counts[TypeNote] != 0 {
		t.Errorf("expected no note count, got %d", shape.Counts[TypeNote])
	}
	if len(shape.Entries) != 3 {
		t.Errorf("entries = %d, want 3", len(shape.Entries))
	}
}

func TestSearch(t *testing.T) {
	s := seedProject(t)
	_, _ = s.Write("p", TypeDecision, "We chose SendGrid for transactional email.", WriteOpts{Name: "d1"})
	hits, err := s.Search("sendgrid")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("got %d hits, want 1", len(hits))
	}
	if hits[0].RelPath != "projects/p/decisions/d1.md" {
		t.Errorf("relPath = %q", hits[0].RelPath)
	}
	if hits[0].Snippet == "" {
		t.Errorf("empty snippet")
	}
}

func TestOrderingKeyMonotonic(t *testing.T) {
	prev := ""
	for i := range 2000 {
		k := OrderingKey()
		if k <= prev {
			t.Fatalf("ordering key not strictly increasing at %d: %q <= %q", i, k, prev)
		}
		prev = k
	}
}

func names(es []Entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Name
	}
	return out
}
