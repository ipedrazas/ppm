package memory

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CollectionTypes lists the collection entry types in a stable display order.
var CollectionTypes = []EntryType{
	TypeDecision, TypeQuestion, TypeTask, TypeNote, TypeConversation,
}

// Title returns a short, heading-stripped first line for an entry body.
func Title(body string) string { return firstLine(body) }

// List returns all entries of a collection type for a project, newest first.
// A missing subdirectory yields an empty slice rather than an error.
func (s *Store) List(project string, t EntryType) ([]Entry, error) {
	reg, ok := Registries[t]
	if !ok || reg.Cardinality != Collection {
		return []Entry{}, nil
	}
	sub := filepath.Join(s.projectDir(project), reg.At)
	ents, err := os.ReadDir(sub)
	if err != nil {
		return []Entry{}, nil
	}
	out := make([]Entry, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(sub, e.Name()))
		if err != nil {
			return nil, err
		}
		fm, body := ParseDoc(string(raw))
		out = append(out, Entry{
			Project:     project,
			Type:        t,
			Name:        strings.TrimSuffix(e.Name(), ".md"),
			Frontmatter: fm.ToMap(),
			Body:        body,
			RelPath:     projectRel(project, reg.At, e.Name()),
		})
	}
	// Newest first by ts, falling back to created date then name.
	sort.SliceStable(out, func(i, j int) bool {
		return sortKey(out[i]) > sortKey(out[j])
	})
	return out, nil
}

func sortKey(e Entry) string {
	if v := fmString(e.Frontmatter, "ts"); v != "" {
		return v
	}
	if v := fmString(e.Frontmatter, "created"); v != "" {
		return v
	}
	return e.Name
}

func fmString(fm map[string]any, key string) string {
	if v, ok := fm[key].(string); ok {
		return v
	}
	return ""
}

// Recent returns at most n entries of a type, newest first. A negative n means
// "all".
func (s *Store) Recent(project string, t EntryType, n int) ([]Entry, error) {
	es, err := s.List(project, t)
	if err != nil {
		return nil, err
	}
	if n >= 0 && len(es) > n {
		es = es[:n]
	}
	return es, nil
}

// OpenQuestions returns the project's unresolved questions (status open or
// absent), newest first.
func (s *Store) OpenQuestions(project string) ([]Entry, error) {
	es, err := s.List(project, TypeQuestion)
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(es))
	for _, e := range es {
		status := fmString(e.Frontmatter, "status")
		if status == "" || status == "open" {
			out = append(out, e)
		}
	}
	return out, nil
}

// Shape returns a project's inventory: counts plus per-entry titles and dates,
// without loading content into the result beyond the derived title.
func (s *Store) Shape(project string) (*ProjectShape, error) {
	idxRaw, _ := s.Read(project, "", "")
	idxFM, _ := ParseDoc(idxRaw)

	counts := map[EntryType]int{}
	var entries []ShapeLine
	for _, t := range CollectionTypes {
		es, err := s.List(project, t)
		if err != nil {
			return nil, err
		}
		if len(es) > 0 {
			counts[t] = len(es)
		}
		for _, e := range es {
			title := Title(e.Body)
			if title == "" {
				title = e.Name
			}
			entries = append(entries, ShapeLine{
				Type:   t,
				Name:   e.Name,
				Title:  title,
				Date:   fmString(e.Frontmatter, "created"),
				Status: fmString(e.Frontmatter, "status"),
			})
		}
	}
	return &ProjectShape{
		Project: project,
		Title:   orElse(idxFM, "title", project),
		Status:  orElse(idxFM, "status", "active"),
		Counts:  counts,
		Entries: entries,
	}, nil
}

func orElse(fm Frontmatter, key, def string) string {
	if v, ok := fm.Get(key); ok && v != "" {
		return v
	}
	return def
}

// SearchHit is a single full-text match with its provenance.
type SearchHit struct {
	RelPath string `json:"relPath"`
	Snippet string `json:"snippet"`
}

// Search does a naive case-insensitive full-text scan across all .md files under
// the memory root, returning matches with a snippet and root-relative path.
func (s *Store) Search(query string) ([]SearchHit, error) {
	q := strings.ToLower(query)
	var hits []SearchHit
	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		i := strings.Index(strings.ToLower(string(raw)), q)
		if i < 0 {
			return nil
		}
		start := max(0, i-40)
		end := min(len(raw), i+80)
		snippet := strings.TrimSpace(strings.ReplaceAll(string(raw[start:end]), "\n", " "))
		rel, _ := filepath.Rel(s.root, path)
		hits = append(hits, SearchHit{RelPath: filepath.ToSlash(rel), Snippet: snippet})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return hits, nil
}
