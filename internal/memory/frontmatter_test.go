package memory

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseDocNoFrontmatter(t *testing.T) {
	fm, body := ParseDoc("# Just a heading\n\nsome text\n")
	if !fm.IsEmpty() {
		t.Errorf("expected empty frontmatter, got %v", fm.ToMap())
	}
	if body != "# Just a heading\n\nsome text" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestParseSerializeRoundTrip(t *testing.T) {
	src := "---\ntype: question\nstatus: open\ncreated: \"2026-06-25\"\n---\n\nDo we have funnel analytics?"
	fm, body := ParseDoc(src)
	if got, _ := fm.Get("status"); got != "open" {
		t.Errorf("status = %q, want open", got)
	}
	if body != "Do we have funnel analytics?" {
		t.Errorf("body = %q", body)
	}
	out := SerializeDoc(fm, body)
	fm2, body2 := ParseDoc(out)
	if got, _ := fm2.Get("type"); got != "question" {
		t.Errorf("round-trip type = %q", got)
	}
	if body2 != body {
		t.Errorf("round-trip body = %q want %q", body2, body)
	}
}

func TestSetPreservesOrderAndReplacesInPlace(t *testing.T) {
	fm := NewFrontmatter()
	fm.Set("a", "1")
	fm.Set("b", "2")
	fm.Set("a", "3") // replace in place, must not move to the end
	out := SerializeDoc(fm, "body")
	ai := strings.Index(out, "a: ")
	bi := strings.Index(out, "b: ")
	if ai < 0 || bi < 0 || ai > bi {
		t.Errorf("expected a before b, got:\n%s", out)
	}
	if got, _ := fm.Get("a"); got != "3" {
		t.Errorf("a = %q, want 3", got)
	}
}

func TestNestedFrontmatterPreserved(t *testing.T) {
	fm := NewFrontmatter()
	fm.Set("id", "onboarding")
	tracker := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
		scalarNode("system"), scalarNode("linear"),
		scalarNode("project"), scalarNode("Onboarding"),
	}}
	fm.SetNode("tracker", tracker)

	out := SerializeDoc(fm, "# Onboarding")
	fm2, _ := ParseDoc(out)
	m := fm2.ToMap()
	tr, ok := m["tracker"].(map[string]any)
	if !ok {
		t.Fatalf("tracker not a nested map: %#v", m["tracker"])
	}
	if tr["system"] != "linear" || tr["project"] != "Onboarding" {
		t.Errorf("nested tracker not preserved: %#v", tr)
	}
}
