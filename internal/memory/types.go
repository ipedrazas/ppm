// Package memory implements the directory-per-project memory format described
// in plans/memory-format.md. It is a Go port of the reference data/store.ts.
package memory

// EntryType is one member of the closed set of memory entry kinds. The write
// path refuses anything not registered here — there is no free-form file write.
type EntryType string

const (
	TypeIndex        EntryType = "index"
	TypeSummary      EntryType = "summary"
	TypeFocus        EntryType = "focus"
	TypeDecision     EntryType = "decision"
	TypeQuestion     EntryType = "question"
	TypeTask         EntryType = "task"
	TypeNote         EntryType = "note"
	TypeConversation EntryType = "conversation"

	// TypeStandard is a workspace-level cross-cutting invariant. It lives in the
	// workspace standards/ collection (parallel to projects/), not under a project,
	// so it is addressed by dedicated store methods rather than the project-scoped
	// Write. See WorkspaceRegistries.
	TypeStandard EntryType = "standard"

	// TypeWaiver is a project-scoped, reasoned exception of one standard. It is a
	// normal project collection (so it flows through Write), but is deliberately
	// kept out of CollectionTypes — it is a cross-cutting overlay consumed by
	// audit, not part of a project's narrative inventory.
	TypeWaiver EntryType = "waiver"

	// TypeInitiative is a workspace-level cross-project campaign (a GDPR review, a
	// migration). Like a standard it lives in a workspace collection; member work
	// is a normal task entry in each project that backlinks to the initiative.
	TypeInitiative EntryType = "initiative"

	// TypeVerdict is a project-scoped record of a manual standard's judgement
	// (pass/fail + rationale), letting a manual standard resolve beyond "unknown".
	// Like waiver it is a project collection kept out of CollectionTypes.
	TypeVerdict EntryType = "verdict"
)

// Cardinality distinguishes singletons (one file at the project root) from
// collections (many entries in a typed subdirectory).
type Cardinality int

const (
	Singleton Cardinality = iota
	Collection
)

// Registry describes where an entry type lives on disk.
type Registry struct {
	Cardinality Cardinality
	// At is the file name for singletons or the subdirectory for collections.
	At string
}

// Registries is the source of truth for the closed entry-type set. Filenames and
// folders are human-friendly convention; the type in frontmatter is canonical.
var Registries = map[EntryType]Registry{
	TypeIndex:        {Singleton, "index.md"},
	TypeSummary:      {Singleton, "summary.md"},
	TypeFocus:        {Singleton, "focus.md"},
	TypeDecision:     {Collection, "decisions"},
	TypeQuestion:     {Collection, "questions"},
	TypeTask:         {Collection, "tasks"},
	TypeNote:         {Collection, "notes"},
	TypeConversation: {Collection, "conversations"},
	TypeWaiver:       {Collection, "waivers"},
	TypeVerdict:      {Collection, "verdicts"},
}

// WorkspaceRegistries is the source of truth for workspace-level collections —
// cross-cutting concerns that belong to no single project and live at the memory
// root (parallel to projects/). They are addressed by dedicated store methods,
// keeping the project-scoped Write path unchanged.
var WorkspaceRegistries = map[EntryType]Registry{
	TypeStandard:   {Collection, "standards"},
	TypeInitiative: {Collection, "initiatives"},
}

// Entry is a single memory entry with its parsed frontmatter and body.
type Entry struct {
	Project     string         `json:"project"`
	Type        EntryType      `json:"type"`
	Name        string         `json:"name"`
	Frontmatter map[string]any `json:"frontmatter,omitempty"`
	Body        string         `json:"body"`
	// RelPath is the slash-separated path relative to the memory root,
	// e.g. projects/onboarding/decisions/2026-06-25-x.md.
	RelPath string `json:"relPath"`
}

// ShapeLine is one entry in a project's inventory — orientation without content.
type ShapeLine struct {
	Type   EntryType `json:"type"`
	Name   string    `json:"name"`
	Title  string    `json:"title"`
	Date   string    `json:"date,omitempty"`
	Status string    `json:"status,omitempty"`
}

// ProjectShape is the inventory of a project: counts plus per-entry titles and
// dates, readable without opening any content.
type ProjectShape struct {
	Project string            `json:"project"`
	Title   string            `json:"title"`
	Status  string            `json:"status"`
	Counts  map[EntryType]int `json:"counts"`
	Entries []ShapeLine       `json:"entries"`
}
