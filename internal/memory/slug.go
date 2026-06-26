package memory

import "strings"

const maxSlugLen = 60

// Slugify lowercases s, collapses runs of non-alphanumerics into single
// hyphens, trims hyphens, and caps the length. Empty input yields "entry".
// Ported from the slugify in data/store.ts.
func Slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > maxSlugLen {
		out = strings.Trim(out[:maxSlugLen], "-")
	}
	if out == "" {
		return "entry"
	}
	return out
}
