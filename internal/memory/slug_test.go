package memory

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Onboarding Drop-off":    "onboarding-drop-off",
		"  spaces  everywhere  ": "spaces-everywhere",
		"Foo/Bar_Baz!!":          "foo-bar-baz",
		"":                       "entry",
		"---":                    "entry",
		"ENG-123":                "eng-123",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSlugifyCapsLength(t *testing.T) {
	long := strings.Repeat("a", 100)
	if got := Slugify(long); len(got) != maxSlugLen {
		t.Errorf("len = %d, want %d", len(got), maxSlugLen)
	}
}
