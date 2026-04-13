# your-ai-memory

A system for maintaining multiple small, focused LLM wikis — each a fully
isolated Claude Code project with its own `CLAUDE.md`. A central hub (TUI +
npm CLI) scaffolds new wikis and enables cross-wiki queries, while keeping each
individual Claude session's context window small.

---

## Why

A single big knowledge base quickly overwhelms Claude's context window. Instead,
**your-ai-memory** gives you a fleet of tiny, domain-focused wikis. When you
open a wiki in Claude Code, Claude only loads the knowledge for *that* domain —
not your entire library. The TUI hub lets you navigate, search, and query across
all wikis from one place.

---

## Quick start

```bash
# Scaffold a new wiki (no installation needed)
npx your-ai-memory create

# Open an existing wiki in Claude Code
npx your-ai-memory open ml-llms

# Launch the TUI hub
npx your-ai-memory hub
# or, after go install:
your-ai-memory-hub
```

Inside any wiki (with Claude Code open):

```
/ingest articles/my-article.md   ← extract knowledge into wiki pages
/update-overview                 ← regenerate wiki/overview.md
```

---

## Architecture

```
~/.your-ai-memory/
└── config.json           ← global registry of all wikis

~/wikis/
├── ml-llms/              ← isolated Claude Code project
│   ├── CLAUDE.md         ← domain identity + protocols
│   ├── .claude/commands/
│   │   └── ingest.md     ← /ingest slash command
│   ├── wiki/
│   │   ├── index.md
│   │   ├── log.md
│   │   ├── overview.md
│   │   ├── concepts/
│   │   ├── patterns/
│   │   ├── papers/
│   │   ├── people/
│   │   ├── connections/
│   │   └── questions/
│   └── raw/
│       ├── articles/
│       └── papers/
│
├── python/               ← another isolated project
└── wine/
```

Two installed components:

| Component | What it does |
|-----------|-------------|
| **npm CLI** (`npx your-ai-memory`) | Scaffold wikis, manage registry, open in Claude Code |
| **Go TUI** (`your-ai-memory-hub`) | Visual hub: browse, search, add sources, run ops |

---

## npm CLI reference

```bash
your-ai-memory                  # interactive menu
your-ai-memory create           # scaffold a new wiki (7-question flow)
your-ai-memory list             # list all registered wikis, grouped by topic
your-ai-memory open [name]      # open wiki in Claude Code
your-ai-memory hub              # launch Go TUI hub
```

The `create` flow asks:

1. Wiki name (becomes the directory name)
2. Location (default: `~/wikis/<name>`)
3. Domain — the topic this wiki covers
4. Goal — learning / research / professional reference / expertise-sharing
5. Source types — articles, papers, videos, books, notes, docs
6. Topic tags — for grouping and cross-wiki search (e.g. `computer-science`)
7. Language — for generated wiki pages

After answering, it scaffolds the directory, writes a domain-configured
`CLAUDE.md`, and registers the wiki in `~/.your-ai-memory/config.json`.

---

## TUI hub keybindings

### Hub (home)

| Key | Action |
|-----|--------|
| `↑ ↓` | Navigate wiki list |
| `enter` | Open wiki detail |
| `n` | Create new wiki |
| `/` | Cross-wiki search |
| `q` | Quit |

### Wiki detail

| Key | Action |
|-----|--------|
| `o` | Open in Claude Code |
| `s` | Search this wiki |
| `a` | Add source (URL) |
| `i` | Run `/ingest` |
| `l` | Operations menu |
| `esc` | Back to hub |

### Search

| Key | Action |
|-----|--------|
| type | Live fuzzy search |
| `tab` | Toggle single / cross-wiki |
| `↑ ↓` | Navigate results |
| `p` | Toggle preview pane |
| `c` | Ask Claude (assemble context) |
| `enter` | Open file |
| `esc` | Back |

### Sources (`raw/` browser)

| Key | Action |
|-----|--------|
| `a` | Add URL stub |
| `m` | Mark for ingestion |
| `esc` | Back |

---

## How INGEST works

1. Drop a source file into `raw/` (article, paper, video transcript, etc.)
2. Open the wiki in Claude Code: `cd ~/wikis/ml-llms && claude`
3. Run `/ingest articles/my-article.md`
4. Claude reads the file, extracts knowledge units, and creates/updates pages
   in the appropriate `wiki/` subdirectory
5. `wiki/index.md` and `wiki/log.md` are updated automatically

Knowledge stays scoped — Claude only sees this wiki's `CLAUDE.md` and pages,
not your other wikis.

---

## How cross-wiki search works

Context is intentionally bounded — only catalog files, not full page content:

1. Select a topic tag (e.g. `computer-science`) in the TUI hub
2. The TUI collects `wiki/index.md` and `wiki/overview.md` from every wiki with
   that tag
3. It assembles a context block:
   ```
   ## Wiki: ml-llms (Machine Learning and LLMs)
   ### index.md
   ...
   ### overview.md
   ...

   ## Wiki: python (Python programming)
   ...

   **Query:** how do transformers relate to Python generators?
   ```
4. Press `c` — the TUI writes the assembled context to a temp file and opens
   `claude -p <tempfile>`, giving Claude the cross-wiki summary plus your query

---

## Install the TUI

Requires Go 1.22+:

```bash
go install github.com/nicolaieilstrup/your-ai-memory/tui@latest
```

Or build locally:

```bash
cd tui
go mod tidy
go build -o your-ai-memory-hub .
```

---

## Repository layout

```
your-ai-memory/
├── README.md
├── cli/                    ← npm package
│   ├── package.json
│   ├── bin/your-ai-memory.js
│   └── lib/
│       ├── create.js
│       ├── list.js
│       ├── open.js
│       └── registry.js
├── tui/                    ← Go TUI hub
│   ├── main.go
│   ├── go.mod
│   ├── app/                ← Bubbletea screens
│   └── wiki/               ← registry, reader, search
└── template/               ← wiki scaffold templates
    ├── CLAUDE.md.tmpl
    ├── .claude/commands/ingest.md
    └── wiki/ + raw/
```

---

## License

MIT
