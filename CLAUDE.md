# webui

This application allows the user to maintain local files hosted on
disk through a web ui. The purpose is to turn local files into
reports and forms allowing on disk files to store the information.

The application assumes it is started at the root of a git repo.
Within that repo a file called `.webui.yaml` will be expected:

```yaml
server:
  port: 3110
  index: README.md

browser:
  application: firefox
  autostart: true

content:
  - name: "help"
    path: help/
    content: markdown
    serve_images: true
    menu: "Help"

data:
  - name: "contacts"
    path: data/contacts.csv
    schema: schema/contacts.yml
    menu: "Data / Contacts"
```

Content sections are folders of files rendered as readable pages (markdown or raw).
Data sections are CSV files exposed as browsable, editable tables (CRUD).

## Project Structure

```
cmd/webui/          entry point — subcommand dispatch (init, serve, version)
internal/config/    YAML config loader, defaults writer, path validator
internal/content/   HTTP handler for markdown/static content folders
internal/data/      HTTP handler for CSV data tables (list, edit, save)
internal/schema/    YAML schema loader (field types, labels, options)
internal/server/    net/http server with graceful shutdown
internal/tmpl/      html/template renderer (clone-per-page, nav builder)
static/             //go:embed tree: css, js, templates
  css/bootstrap.min.css   Bootstrap 5.3.3 (bundled, MIT)
  css/style.css            app-specific overrides (markdown, cards, overlay)
  js/bootstrap.bundle.min.js
  js/main.js               active-link detection, confirm dialogs, stopServer()
  templates/base.html      Bootstrap navbar with dropdown support
  templates/index.html     dashboard card grid
  templates/data.html      data table (table-hover, table-striped)
  templates/form.html      edit form (form-control, form-select)
  templates/content.html   markdown page
utils/              FileExists, AbsPath helpers
example/            self-contained example repo (see below)
.github/workflows/  CI and release workflows
LICENSE             project MIT + Bootstrap 5.3.3 MIT attribution
```

## Key Implementation Details

### Template rendering
`html/template` does not support block overrides across a shared template set.
Each page uses a **clone-per-page** strategy: `base.html` is parsed first, then
each page template is applied to a fresh clone so `{{define "content"}}` overrides
`{{block "content"}}` correctly.

### Navigation
Nav is built at startup from the config and stored on the renderer. Content
handlers call `StaticNav()` (returns a copy) and `RenderWithNav()` to inject a
per-request enriched nav (directory file listing as dropdown children).

### CSV reading (`internal/data/handler.go`)
`readCSV()` handles all real-world CSV quirks:
- Strips UTF-8 BOM (Excel exports) via `bufio.Reader.Peek/Discard`
- `FieldsPerRecord = -1` — tolerates rows with fewer columns than the header
- `LazyQuotes = true` — accepts unescaped quotes in field values
- `strings.TrimSpace` on every cell — removes trailing `\r` from CRLF files

### Atomic CSV write
`writeCSV()` creates a temp file in the same directory then uses `os.Rename`
to replace the original. `defer os.Remove(tmpName)` is a no-op after a
successful rename.

### Bootstrap dropdowns
The custom CSS-only hover dropdown was replaced with Bootstrap 5.3.3's native
`data-bs-toggle="dropdown"` mechanism, which works correctly in all browsers.
Bootstrap JS (Popper included) is embedded in `static/js/bootstrap.bundle.min.js`.

### Version injection
`cmd/webui/main.go` declares `var version = "dev"`. The release workflow
overwrites it at link time:
```
go build -ldflags="-X main.version=v1.2.3" ./cmd/webui
```

## Example directory

`example/` is a self-contained demo repo. Run it with:
```bash
cd example
webui serve
```

Contents:
- `.webui.yaml` — 1 content section (help), 3 data sections
- `help/` — markdown docs (index, getting-started, configuration)
- `data/contacts.csv` — 9 rows, 8 columns
- `data/projects.csv` — 10 rows, 8 columns
- `data/inventory.csv` — 12 rows, 8 columns (demonstrates quoted CSV: `"27"" Monitor"`)
- `schema/contacts.yml` / `projects.yml` / `inventory.yml` — typed schemas with readonly IDs, select dropdowns, required fields

## Development Guidelines

### Coding Style

- Keep the code clean and simple; split large blocks into focused functions
- Add doc comments on exported symbols; avoid obvious inline comments
- Ensure all code is testable; prefer interfaces at package boundaries
- No half-finished implementations; no unused backwards-compat shims

### Testing Approach

- Write tests alongside implementation (TDD where practical)
- Test edge cases and error conditions — especially CSV parsing quirks
- Use `httptest.NewRecorder` for HTTP handler tests
- Regression tests live in the same `_test.go` file as the feature

### Files to Avoid

- `node_modules/`, `.env` files
- Build artifacts (`dist/`, `build/`)
- Log files

## CI / Release

- `.github/workflows/ci.yml` — `go vet` + `go test -race` on Go 1.22 and 1.23,
  triggered on push to `main` and all PRs
- `.github/workflows/release.yml` — cross-compiles for linux/amd64, linux/arm64,
  darwin/amd64, darwin/arm64, windows/amd64 on `v*` tag push; uploads binaries
  and SHA-256 checksums; creates a GitHub release with auto-generated notes

## Working with Claude Code

### Preferred Workflow

1. **Explore** — read the relevant files before changing anything
2. **Plan** — discuss the approach for non-trivial changes
3. **Code** — implement with tests; run `go test ./...` before reporting done
4. **Commit** — use conventional commits (`feat:`, `fix:`, `docs:`, etc.)

### Git Workflow

- Branch naming: `feature/*`, `bugfix/*`, `docs/*`
- Prefer one focused commit per logical change

### Communication Style

- Ask clarifying questions before implementing ambiguous requests
- Explain non-obvious technical decisions
- Point out potential issues proactively
- Keep responses concise; avoid restating what the diff already shows
