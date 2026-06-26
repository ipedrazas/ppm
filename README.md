# ppm — Project & Product Manager CLI

`ppm` manages the markdown memory system for a PM / Product-Owner agent: a
directory-per-project tree of **typed** entries (decisions, questions, tasks,
notes, conversations) plus per-project `index` / `summary` / `focus` singletons.

The format is plain Markdown with YAML frontmatter — drop the `memory/` folder
into Obsidian and it renders. The anti-dumping-ground guarantee is structural:
every entry has a **type from a closed set** and there is no free-form "write any
file" command. See [`plans/memory-format.md`](plans/memory-format.md) for the
full format spec.

Output is **JSON by default** (the CLI is meant to be driven by an agent); pass
`-o text` (or `--pretty`) for human-readable output.

## Install

```sh
go build -o ppm .
# or with a version stamp:
go build -ldflags "-X github.com/ipedrazas/ppm/cmd.version=v0.1.0" -o ppm .
```

## Memory root resolution

The memory root is chosen in this order:

1. `--root <dir>` flag
2. `$PPM_MEMORY_ROOT`
3. the nearest ancestor of the cwd containing an existing `memory/` directory
4. default `./memory`

## Quick start

```sh
ppm init                                   # scaffold the workspace
ppm project create onboarding --title "Onboarding drop-off"

ppm decision add onboarding --content "Email nudge first; cheap and testable."
ppm question add onboarding --name funnel --content "Do funnel analytics exist?"
ppm question resolve onboarding funnel --content "Yes — no new instrumentation."
ppm task add onboarding --ref ENG-123 --url https://linear.app/acme/issue/ENG-123 \
    --content "Onboarding email nudge. Scope: email only."

ppm summary set onboarding --content "Reduce onboarding drop-off via nudges."
ppm focus   set onboarding --content "Shipping the email nudge (ENG-123)."

ppm project show onboarding                 # shape: inventory without content
ppm search "funnel"                         # full-text search with provenance
ppm context onboarding                      # the shape-aware injected slice
```

## Content input

Commands that take a body accept `--content` (primary) or `--file <path>`
(fallback). Exactly one must be given.

## Commands

| Command | Purpose |
|---|---|
| `ppm init` | Scaffold `index.md`, `preferences.md`, `glossary.md`, `projects/` |
| `ppm project create <slug> --title T` | Create a project (scaffolds index/summary/focus) |
| `ppm project list` | List all projects |
| `ppm project show <slug>` | Project shape (entry inventory, no content) |
| `ppm project update <slug> [--status\|--title\|--tracker-*]` | Edit index frontmatter |
| `ppm read [project] [--type T] [--name N]` | Full content (no project → workspace index) |
| `ppm search <query>` | Full-text search across all memory |
| `ppm context <project> [--recent N]` | Emit the injected context slice |
| `ppm decision add <project> [--name]` | Record a dated decision + rationale |
| `ppm decision list <project> [--recent N]` | List decisions (newest first) |
| `ppm question add <project> [--name]` | Record an open question |
| `ppm question resolve <project> <name>` | Flip a question to resolved |
| `ppm question list <project> [--open]` | List questions |
| `ppm task add <project> --ref R [--url]` | Add a task reference + rationale |
| `ppm task list <project>` | List tasks |
| `ppm note add <project> [--name]` | Add a note |
| `ppm conversation add <project> [--name]` | Add a conversation (alias `conv`) |
| `ppm summary set <project>` | Replace the project summary |
| `ppm focus set <project>` | Replace the project focus |

Global flags: `--root`, `-o/--output json|text`, `--pretty`, `--version`.

## Output contract

Every command emits a uniform envelope. JSON:

```json
{ "ok": true, "message": "…", "data": { /* structured payload */ } }
```

Errors set `"ok": false` with an `"error"` field and a non-zero exit code. In
JSON mode the error envelope is written to **stdout** (uniform parsing); in text
mode it is written to **stderr**.

## Design notes

- **Type in frontmatter is canonical**; folders and filenames are convention.
- **`ts` ordering** uses UUIDv7 — time-sortable and monotonic across separate
  CLI invocations, so rapid writes stay correctly ordered.
- **Frontmatter** is real YAML (key order and nested `tracker` preserved).
- **Shape vs content**: the entry inventory is first-class signal, readable
  without opening any entry; `context` injects full content only for the cheap,
  high-value entries and shape-only for the rest.

## Development

```sh
go build ./...
go vet ./...
go test ./...
```
