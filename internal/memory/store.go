package memory

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

// Store is the memory root and all operations over it.
type Store struct {
	root string
}

// NewStore returns a Store rooted at root (an absolute path is expected).
func NewStore(root string) *Store { return &Store{root: root} }

// Root returns the absolute memory root path.
func (s *Store) Root() string { return s.root }

func (s *Store) projectsDir() string        { return filepath.Join(s.root, "projects") }
func (s *Store) projectDir(p string) string { return filepath.Join(s.projectsDir(), p) }

// projectRel builds a slash-separated, root-relative path under projects/.
func projectRel(parts ...string) string {
	return path.Join(append([]string{"projects"}, parts...)...)
}

func (s *Store) abs(rel string) string {
	return filepath.Join(s.root, filepath.FromSlash(rel))
}

// Init scaffolds the workspace: projects/ plus the three workspace singletons,
// each created only if absent.
func (s *Store) Init() error {
	if err := os.MkdirAll(s.projectsDir(), dirPerm); err != nil {
		return err
	}
	for _, f := range []string{"index.md", "preferences.md", "glossary.md"} {
		p := filepath.Join(s.root, f)
		if _, err := os.Stat(p); err != nil {
			name := strings.TrimSuffix(f, ".md")
			if err := os.WriteFile(p, []byte("# "+name+"\n"), filePerm); err != nil {
				return err
			}
		}
	}
	return nil
}

// ListProjects returns the project slugs, sorted. A missing projects/ dir yields
// an empty slice rather than an error.
func (s *Store) ListProjects() ([]string, error) {
	ents, err := os.ReadDir(s.projectsDir())
	if err != nil {
		return []string{}, nil
	}
	out := make([]string, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	// os.ReadDir already returns names sorted.
	return out, nil
}

// CreateProject scaffolds a new project's index/summary/focus. It errors if the
// project already exists. Returns the resolved slug.
func (s *Store) CreateProject(project, title string) (string, error) {
	slug := Slugify(project)
	dir := s.projectDir(slug)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, "index.md")); err == nil {
		return "", memErrf("project %q already exists", slug)
	}

	idxFM := NewFrontmatter()
	idxFM.Set("id", slug)
	idxFM.Set("title", title)
	idxFM.Set("status", "active")
	idxFM.Set("created", Today())
	idxFM.Set("updated", Today())
	if _, err := s.Write(slug, TypeIndex, "# "+title+"\n", WriteOpts{FM: &idxFM}); err != nil {
		return "", err
	}
	if _, err := s.Write(slug, TypeSummary, "_To be written._", WriteOpts{}); err != nil {
		return "", err
	}
	if _, err := s.Write(slug, TypeFocus, "_No current focus yet._", WriteOpts{}); err != nil {
		return "", err
	}
	return slug, nil
}

// WriteMode selects a non-default mutation for collection writes.
type WriteMode string

const (
	// ModeResolve flips a question's status to resolved.
	ModeResolve WriteMode = "resolve"
)

// KV is an ordered frontmatter key/value pair for collection creates.
type KV struct{ Key, Val string }

// WriteOpts carries the optional parameters of Write.
type WriteOpts struct {
	// Name is the explicit entry slug (required for resolve; optional otherwise).
	Name string
	// Mode selects resolve; empty means a normal create/replace.
	Mode WriteMode
	// Extra are additional frontmatter pairs for collection creates (e.g. a
	// task's ref/url), appended after type/created/ts in order.
	Extra []KV
	// FM is a full frontmatter override for singletons (used for the index).
	FM *Frontmatter
}

// Write is the type-addressable write. Singletons replace; collections create a
// new entry unless Mode is resolve. Unknown types are rejected.
func (s *Store) Write(project string, t EntryType, content string, opts WriteOpts) (*Entry, error) {
	reg, ok := Registries[t]
	if !ok {
		return nil, memErrf("unknown entry type %q", t)
	}
	if err := os.MkdirAll(s.projectDir(project), dirPerm); err != nil {
		return nil, err
	}

	if reg.Cardinality == Singleton {
		return s.writeSingleton(project, t, content, opts.FM)
	}

	sub := filepath.Join(s.projectDir(project), reg.At)
	if err := os.MkdirAll(sub, dirPerm); err != nil {
		return nil, err
	}

	if opts.Mode == ModeResolve {
		return s.resolveQuestion(project, reg, opts.Name, content)
	}
	return s.createCollectionEntry(project, t, reg, content, opts)
}

func (s *Store) writeSingleton(project string, t EntryType, content string, fm *Frontmatter) (*Entry, error) {
	reg := Registries[t]
	rel := projectRel(project, reg.At)
	var f Frontmatter
	if fm != nil {
		f = *fm
	} else {
		f = NewFrontmatter()
	}
	if t != TypeIndex {
		f.Set("type", string(t))
	}
	if err := os.WriteFile(s.abs(rel), []byte(SerializeDoc(f, content)), filePerm); err != nil {
		return nil, err
	}
	return &Entry{
		Project:     project,
		Type:        t,
		Name:        string(t),
		Frontmatter: f.ToMap(),
		Body:        strings.TrimSpace(content),
		RelPath:     rel,
	}, nil
}

func (s *Store) resolveQuestion(project string, reg Registry, name, content string) (*Entry, error) {
	if name == "" {
		return nil, memErrf("resolve requires a name")
	}
	rel := projectRel(project, reg.At, name+".md")
	abs := s.abs(rel)
	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	fm, body := ParseDoc(string(raw))
	fm.Set("status", "resolved")
	newBody := content
	if newBody == "" {
		newBody = body
	}
	if err := os.WriteFile(abs, []byte(SerializeDoc(fm, newBody)), filePerm); err != nil {
		return nil, err
	}
	return &Entry{
		Project:     project,
		Type:        TypeQuestion,
		Name:        name,
		Frontmatter: fm.ToMap(),
		Body:        strings.TrimSpace(newBody),
		RelPath:     rel,
	}, nil
}

func (s *Store) createCollectionEntry(project string, t EntryType, reg Registry, content string, opts WriteOpts) (*Entry, error) {
	name := opts.Name
	if name != "" {
		name = Slugify(name)
	} else {
		short := strings.Join(firstWords(firstLine(content), 7), " ")
		name = Today() + "-" + Slugify(short)
	}

	fm := NewFrontmatter()
	fm.Set("type", string(t))
	fm.Set("created", Today())
	fm.Set("ts", OrderingKey())
	for _, kv := range opts.Extra {
		fm.Set(kv.Key, kv.Val)
	}
	if t == TypeQuestion {
		if _, ok := fm.Get("status"); !ok {
			fm.Set("status", "open")
		}
	}

	rel := projectRel(project, reg.At, name+".md")
	if err := os.WriteFile(s.abs(rel), []byte(SerializeDoc(fm, content)), filePerm); err != nil {
		return nil, err
	}
	return &Entry{
		Project:     project,
		Type:        t,
		Name:        name,
		Frontmatter: fm.ToMap(),
		Body:        strings.TrimSpace(content),
		RelPath:     rel,
	}, nil
}

// Read returns raw file content. With no project, the workspace index; with a
// project but no type, the project index; otherwise the singleton file or the
// named collection entry.
func (s *Store) Read(project string, t EntryType, name string) (string, error) {
	if project == "" {
		return readFile(filepath.Join(s.root, "index.md"))
	}
	if t == "" {
		return readFile(filepath.Join(s.projectDir(project), "index.md"))
	}
	reg, ok := Registries[t]
	if !ok {
		return "", memErrf("unknown entry type %q", t)
	}
	if reg.Cardinality == Singleton {
		return readFile(filepath.Join(s.projectDir(project), reg.At))
	}
	if name == "" {
		return "", memErrf("type %q requires a name", t)
	}
	return readFile(filepath.Join(s.projectDir(project), reg.At, name+".md"))
}

func readFile(p string) (string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// firstLine returns the first non-empty line of body with leading markdown
// heading hashes stripped, capped at 80 characters.
func firstLine(body string) string {
	for line := range strings.SplitSeq(body, "\n") {
		stripped := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#"))
		if stripped != "" {
			if len(stripped) > 80 {
				stripped = stripped[:80]
			}
			return stripped
		}
	}
	return ""
}

func firstWords(s string, n int) []string {
	words := strings.Fields(s)
	if len(words) > n {
		words = words[:n]
	}
	return words
}
