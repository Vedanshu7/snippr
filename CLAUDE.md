# Snippr — Claude Code Instructions

## Project
Self-hosted snippet manager. Go · Chi · SQLite (modernc, pure Go) · HTMX.
See `docs/STYLE.md` for the full coding style guide.

## Stack
- Router: `github.com/go-chi/chi/v5`
- Auth: `github.com/go-chi/jwtauth/v5` (HS256 JWT)
- DB: `modernc.org/sqlite` — pure Go, no CGO needed
- Password: `golang.org/x/crypto/bcrypt`
- Config: `github.com/joho/godotenv`

## Layout
```
cmd/server/      web server entrypoint
cmd/cli/         CLI entrypoint (Phase 2)
internal/config/ env loading
internal/db/     data access only — no HTTP, no business logic
internal/models/ plain structs, zero deps
internal/api/    HTTP handlers + middleware
migrations/      root-level (informational only)
internal/db/migrations/  actual embedded migrations (go:embed resolves relative to package)
docs/STYLE.md    full coding style guide
```

## Key Style Rules (enforced — see docs/STYLE.md for full detail)

**Handlers** are always factory functions: `func XxxHandler(db *sql.DB) http.HandlerFunc`

**DB functions** return `nil, nil` for not-found — never surface `sql.ErrNoRows`

**Types**: use `any` not `interface{}`. Use `COALESCE` in SQL, never `sql.NullString`.

**Errors**:
- DB layer: wrap with `fmt.Errorf("operation: %w", err)`, log nowhere
- API layer: `slog.Error(...)` then `writeError(w, status, msg)` then `return`
- Never write to `w` directly — only `writeJSON` / `writeError`

**SQL**: always explicit column list, always `AND user_id = ?`, parameterized only (never `fmt.Sprintf`)

**Handler body order**: Decode → Validate → Authorize → Act → Respond

**Naming**: `any` not `interface{}`. No stutter. Verbs for functions, nouns for types. Acronyms all-caps (`userID`, `parseURL`).

**Comments**: only the WHY. Exported symbols get godoc. `// TODO(name): reason`. Empty lines → section comments.

**Tests**: in-memory SQLite (`:memory:`), no mocking, `t.Helper()` in every helper, table-driven for input variation.

**Never**: `init()`, global mutable state, panics in `internal/`, `fmt.Println`, `sql.NullString`, `SELECT *`.

## Commands
```
make run          start server (port 8080)
make build        compile to bin/snippr
make test         run all tests
make docker-up    start via docker compose
```

## Phase Status
- Phase 1 (Core Backend): complete — auth, CRUD, FTS5, public slugs
- Phase 2: Web UI (html/template + HTMX) + CLI tool
- Phase 3: Stripe billing + team workspaces
