# Snippr — Go Coding Style Guide

> Inspired by [causify-ai/helpers coding style](https://github.com/causify-ai/helpers/blob/master/docs/code_guidelines/all.coding_style.how_to_guide.md) and adapted for Go.

---

<!-- toc -->
- [High-Level Principles](#high-level-principles)
- [Naming](#naming)
- [Comments](#comments)
- [Error Handling](#error-handling)
- [Imports](#imports)
- [Functions and Signatures](#functions-and-signatures)
- [Writing Clear Code](#writing-clear-code)
- [Writing Robust Code](#writing-robust-code)
- [Packages and Architecture](#packages-and-architecture)
- [Logging](#logging)
- [Testing](#testing)
- [SQL](#sql)
- [HTTP Handlers](#http-handlers)
- [Concurrency](#concurrency)
- [Tooling](#tooling)
<!-- tocstop -->

---

## High-Level Principles

These are the axioms everything else derives from. If a lower-level rule feels
wrong in a specific case, come back here and ask which principle applies.

### The Writer Is the Reader

- Code is written once and read hundreds of times — by teammates, by your
  future self, by reviewers.
- Make code easy to read even if that makes it slightly harder to write.
- "It was clear to me when I wrote it" is not a justification.

### DRY — Don't Repeat Yourself

- Every piece of knowledge must have a single, unambiguous representation.
- Duplication is not just copy-paste; it includes structural repetition (two
  functions doing the same job under different names).
- Three lines of similar code: acceptable. Four: extract a helper.

### Least Surprise

- A function named `GetUser` should get a user, not delete one and return it.
- A function that silently swallows errors surprises every caller.
- Follow the conventions the Go standard library sets. Readers already know them.

### End-to-End First

- Implement things end-to-end, then improve each piece.
- A working skeleton with rough edges beats a perfect component with no
  integration.

### Pay the Technical Debt

- Hacks propagate. An ugly workaround today is a three-day refactor next month.
- Leave a `// TODO(name):` when you can't fix it now. Never leave it silent.

### No Ugly Hacks

- We do not tolerate hacks that cost more to undo than doing it right.
- Especially: hidden global state, unexplained `//nolint:`, bypassing auth
  middleware "just for tests", storing secrets in code.

---

## Naming

### Follow Go Conventions

- Go uses `camelCase` for unexported, `CamelCase` for exported identifiers.
- Acronyms are all-caps or all-lower depending on export: `userID`, `UserID`,
  `parseURL`, `ParseURL`. Never `UserId`, `parseUrl`.
- Package names are lowercase, single words, no underscores: `db`, `api`,
  `config`. Never `db_layer`, `apiHandlers`.

### Variables

#### Use Short Names for Short-Lived Variables

- The scope of a variable and the length of its name should be inversely
  proportional. A loop counter is `i`, a file descriptor is `f`.
- But: never sacrifice clarity for brevity in non-trivial scope.
  - _Bad_: `u` for a `*models.User` that lives for 20 lines
  - _Good_: `user` for a `*models.User` that lives for 20 lines

#### Use Descriptive Names at Package Scope

- _Bad_
  ```go
  var d *sql.DB
  ```
- _Good_
  ```go
  var database *sql.DB
  ```

#### Avoid Generic Names Like `data`, `info`, `result`

These names tell you nothing. Name the thing by what it represents.

- _Bad_
  ```go
  data, err := dbpkg.GetSnippet(db, id, userID)
  ```
- _Good_
  ```go
  snippet, err := dbpkg.GetSnippet(db, id, userID)
  ```

#### Do Not Stutter

If the package already communicates the context, do not repeat it in the name.

- _Bad_
  ```go
  // in package db
  func DBOpen(path string) (*sql.DB, error)
  ```
- _Good_
  ```go
  // in package db — the package name provides context
  func Open(path string) (*sql.DB, error)
  ```

### Functions

#### Name Functions by What They Do, Not How

- _Bad_: `runSQLInsertForSnippet` — describes the mechanism
- _Good_: `CreateSnippet` — describes the contract

#### Getter Functions Drop the `Get` Prefix Only for Simple Field Access

- Standard Go style: `user.Name()` not `user.GetName()`.
- For DB access functions that do real work, keep the verb:
  `GetUserByEmail`, `GetSnippetBySlug`.
- Rationale: `GetUserByEmail` signals a DB round-trip; `UserByEmail` looks like
  a pure lookup on a struct.

#### Boolean Functions Use Is, Has, Can, Should

- `IsPublic`, `HasTag`, `CanEdit` — immediately readable in `if` conditions.
  - _Bad_: `if snippet.Public() { ... }`
  - _Good_: `if snippet.IsPublic() { ... }`

### Error Variables

- Package-level sentinel errors use `Err` prefix: `ErrNotFound`, `ErrUnauthorized`.
- Never create sentinel errors you don't actually check with `errors.Is`.

```go
var ErrNotFound = errors.New("not found")
```

### Constants

- Use `camelCase` for unexported, `CamelCase` for exported.
- Group related constants in a `const` block.
- _Bad_
  ```go
  const maxSnippets = 50
  const maxTags = 10
  ```
- _Good_
  ```go
  const (
      maxSnippets = 50
      maxTags     = 10
  )
  ```

### Receivers

- Receiver names are 1-2 letters, derived from the type: `s` for `*Snippet`,
  `u` for `*User`.
- Be consistent: if one method uses `s`, all methods on that type use `s`.
- Never use `self` or `this`.

---

## Comments

### Write Comments That Explain Why, Not What

The code already says what it does. Comments add the reason a reader can't
derive from the code alone.

- _Bad_
  ```go
  // Loop over tags.
  for _, tag := range tags {
  ```
- _Good_
  ```go
  // Tags are trimmed here rather than at insertion to keep the DB layer
  // free of string normalization logic.
  for _, tag := range tags {
      tag = strings.TrimSpace(tag)
  ```

### Exported Symbols Get Godoc Comments

Every exported function, type, and constant must have a godoc comment. The
comment starts with the symbol name.

- _Bad_
  ```go
  // Opens the database and runs migrations.
  func Open(path string) (*sql.DB, error) {
  ```
- _Good_
  ```go
  // Open connects to the SQLite database at path, applies WAL and foreign-key
  // pragmas, and returns the connection. The directory is created if absent.
  func Open(path string) (*sql.DB, error) {
  ```

### Unexported Symbols — Comment Only When Non-Obvious

Do not add godoc to unexported helpers. Only add a `//` comment if the
implementation has a subtle invariant or a non-obvious reason.

- _Bad_
  ```go
  // generateSlug generates a random slug.
  func generateSlug() (string, error) {
  ```
- _Good_
  ```go
  // generateSlug uses crypto/rand to avoid bias: math/rand would not give
  // uniform distribution across the base62 alphabet via modulo.
  func generateSlug() (string, error) {
  ```

### Commenting Out Code

When you comment out code, explain why it is disabled. Dead code without
context is a trap.

- _Bad_
  ```go
  // rows, err = db.Query(`SELECT ... FROM snippets_fts ...`, query)
  rows, err = db.Query(`SELECT ... FROM snippets ...`, userID)
  ```
- _Good_
  ```go
  // TODO(vedanshu): Switch back to FTS once the trigger rebuild is confirmed.
  // rows, err = db.Query(`SELECT ... FROM snippets_fts ...`, query)
  rows, err = db.Query(`SELECT ... FROM snippets ...`, userID)
  ```

### TODO Format

Always include a name so it can be grepped and attributed.

```go
// TODO(vedanshu): Replace with prepared statement when traffic increases.
// TODO(*): Any team member can address this before Phase 2 ships.
```

### Replace Empty Lines With Comments

If you feel the urge to add a blank line inside a function, it usually means
that block deserves a label.

- _Bad_
  ```go
  hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
  if err != nil { ... }

  id, err := dbpkg.CreateUser(db, req.Email, string(hash))
  ```
- _Good_
  ```go
  hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
  if err != nil { ... }
  // Persist the user — email uniqueness is enforced by the DB UNIQUE constraint.
  id, err := dbpkg.CreateUser(db, req.Email, string(hash))
  ```

---

## Error Handling

### The Golden Rule: Never Ignore an Error

Every `err` must be handled. Assigning to `_` is only acceptable for truly
optional side-effects (e.g., closing a body in a deferred call where the error
doesn't matter).

- _Bad_
  ```go
  db.Exec(`DELETE FROM tags WHERE snippet_id = ?`, id)
  ```
- _Good_
  ```go
  if _, err := db.Exec(`DELETE FROM tags WHERE snippet_id = ?`, id); err != nil {
      return err
  }
  ```

### DB Layer: Wrap With Context

Always wrap with `fmt.Errorf("operation: %w", err)` so callers know where the
failure originated.

```go
if err != nil {
    return nil, fmt.Errorf("get snippet: %w", err)
}
```

Use a short verb-noun description matching the function name: `"create user"`,
`"list snippets"`, `"get snippet by slug"`.

### DB Layer: Not Found Is Not an Error

Return `nil, nil` when `sql.ErrNoRows` is encountered. The caller decides
whether "not found" is an error.

- _Bad_
  ```go
  if err == sql.ErrNoRows {
      return nil, ErrNotFound
  }
  ```
- _Good_
  ```go
  if err == sql.ErrNoRows {
      return nil, nil
  }
  ```

The handler then:

```go
snippet, err := dbpkg.GetSnippet(db, id, userID)
if err != nil {
    writeError(w, http.StatusInternalServerError, "server error")
    return
}
if snippet == nil {
    writeError(w, http.StatusNotFound, "snippet not found")
    return
}
```

### API Layer: Write Error Then Return

Call `writeError`, then `return`. Never continue executing after writing a
response. Never call `w.Write` or `w.WriteHeader` directly — use `writeJSON`
and `writeError`.

- _Bad_
  ```go
  if err != nil {
      writeError(w, http.StatusInternalServerError, "server error")
      // no return — execution continues
  }
  ```
- _Good_
  ```go
  if err != nil {
      writeError(w, http.StatusInternalServerError, "server error")
      return
  }
  ```

### API Layer: Generic User-Facing Messages

Never expose internal error messages, SQL text, or stack traces to the caller.

- _Bad_
  ```go
  writeError(w, http.StatusInternalServerError, err.Error())
  ```
- _Good_
  ```go
  slog.Error("create snippet", "err", err)
  writeError(w, http.StatusInternalServerError, "server error")
  ```

### Auth Errors: No Hints

Both "user not found" and "wrong password" return 401 with `"invalid
credentials"`. Do not tell an attacker which one failed.

```go
if user == nil {
    writeError(w, http.StatusUnauthorized, "invalid credentials")
    return
}
if err := bcrypt.CompareHashAndPassword(...); err != nil {
    writeError(w, http.StatusUnauthorized, "invalid credentials")
    return
}
```

### Use `errors.Is` and `errors.As` for Checking

Never compare errors with `==` except for sentinel values in the same package.

- _Bad_
  ```go
  if err == sql.ErrNoRows {
  ```
- _Good_
  ```go
  if errors.Is(err, sql.ErrNoRows) {
  ```
  (Exception: within the `db` package itself, `== sql.ErrNoRows` is acceptable
  since you own the DB code and sql.ErrNoRows is a well-known sentinel.)

---

## Imports

### Three Groups Separated by Blank Lines

1. Standard library
2. Third-party
3. Internal (`github.com/vedanshu/snippr/...`)

```go
import (
    "database/sql"
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/jwtauth/v5"

    dbpkg "github.com/vedanshu/snippr/internal/db"
    "github.com/vedanshu/snippr/internal/models"
)
```

### Alias Only to Avoid Collision

Aliases add cognitive load. Only alias when two packages would have the same
last path element.

```go
dbpkg "github.com/vedanshu/snippr/internal/db"
// — needed because the local var `db *sql.DB` would shadow the package name.
```

### No Dot Imports

`import . "pkg"` pollutes the namespace and hides where each symbol comes from.

- _Bad_
  ```go
  import . "github.com/go-chi/chi/v5"
  ```

---

## Functions and Signatures

### Parameter Order

Inputs first, then outputs (as parameters, not return values), then options.
Dependencies (like `*sql.DB`, `*jwtauth.JWTAuth`) always go first.

```go
// good: dependency → input → input
func CreateSnippet(db *sql.DB, s *models.Snippet) (int64, error)

// bad: inputs before dependency
func CreateSnippet(s *models.Snippet, db *sql.DB) (int64, error)
```

### Keep Related Parameters Grouped

If you always pass `id` and `userID` together, they should be adjacent in the
signature — or consider a small struct if this pattern appears 3+ times.

### Handler Functions Are Factory Functions

All HTTP handlers close over their dependencies. Never use package-level
variables for `*sql.DB` or `*jwtauth.JWTAuth`.

- _Bad_
  ```go
  var db *sql.DB

  func CreateSnippetHandler(w http.ResponseWriter, r *http.Request) { ... }
  ```
- _Good_
  ```go
  func CreateSnippetHandler(db *sql.DB) http.HandlerFunc {
      return func(w http.ResponseWriter, r *http.Request) { ... }
  }
  ```

Rationale: dependency injection via closure makes each handler independently
testable without shared state.

### Avoid Boolean Arguments for Branching Behaviour

A `bool` parameter that completely changes the behaviour of a function is a
sign that you need two functions.

- _Bad_
  ```go
  func ListSnippets(db *sql.DB, userID int64, useFTS bool) ([]models.Snippet, error)
  ```
- _Good_
  ```go
  func ListSnippets(db *sql.DB, userID int64, query, tag string) ([]models.Snippet, error)
  // (empty string = "not set")
  ```

### Return Consistent Types

A function should always return the same type. Returning `nil` for
"no result" is fine; returning `nil` sometimes and a struct sometimes is
confusing.

```go
// good: always (*models.Snippet, error), nil snippet means "not found"
func GetSnippet(db *sql.DB, id, userID int64) (*models.Snippet, error)
```

### Do Not Make Tiny Wrappers

Wrappers that add nothing cost a reader's attention.

- _Bad_
  ```go
  func closeDB(db *sql.DB) {
      db.Close()
  }
  ```
- _Good_: just `defer db.Close()`

### Single Responsibility

If you find yourself naming a function with "and" (`createAndSendEmail`), split
it. Each function does one thing.

---

## Writing Clear Code

### Order Functions in Topological Order

Put helpers above the functions that call them. A reader going top-to-bottom
should not encounter a call to a function defined below.

```
generateSlug()         ← inner helper
nullableSlug()         ← inner helper
CreateSnippet()        ← calls generateSlug, nullableSlug
ListSnippets()
UpdateSnippet()        ← also calls generateSlug
DeleteSnippet()
```

### Distinguish Exported and Unexported

- Exported functions (`CreateSnippet`) form the public API of the package.
- Unexported functions (`generateSlug`, `setTags`) are implementation details.
- Keep unexported helpers close to — typically just above — the exported
  function that uses them.

### Avoid Wall-of-Code Functions

If a function body exceeds ~40 lines or has more than 3 levels of nesting,
extract helpers.

- _Bad_
  ```go
  func ListSnippets(db *sql.DB, userID int64, query, tag string) ([]models.Snippet, error) {
      var rows *sql.Rows
      var err error
      if query != "" {
          rows, err = db.Query(`... FTS query ...`, query, userID)
          if err != nil { return nil, ... }
          defer rows.Close()
          var snippets []models.Snippet
          for rows.Next() {
              // scan ...
              // load tags ...
          }
          return snippets, rows.Err()
      } else if tag != "" {
          // ... entire block repeated
      } else {
          // ... entire block repeated
      }
  }
  ```
- _Good_
  ```go
  func ListSnippets(db *sql.DB, userID int64, query, tag string) ([]models.Snippet, error) {
      rows, err := querySnippetRows(db, userID, query, tag)
      if err != nil {
          return nil, fmt.Errorf("list snippets: %w", err)
      }
      defer rows.Close()
      return scanSnippets(db, rows)
  }
  ```

### Keep Related Code Close

The variable declaration and its first use should be adjacent. Do not declare
all variables at the top of a function (that's C-style, not Go-style).

### Early Returns Over Deep Nesting

Guard clauses at the top, happy path at the bottom.

- _Bad_
  ```go
  func handler(...) {
      if err == nil {
          if user != nil {
              if snippet != nil {
                  writeJSON(w, 200, snippet)
              } else {
                  writeError(w, 404, "not found")
              }
          } else {
              writeError(w, 401, "unauthorized")
          }
      } else {
          writeError(w, 500, "server error")
      }
  }
  ```
- _Good_
  ```go
  func handler(...) {
      if err != nil {
          writeError(w, 500, "server error")
          return
      }
      if user == nil {
          writeError(w, 401, "unauthorized")
          return
      }
      if snippet == nil {
          writeError(w, 404, "not found")
          return
      }
      writeJSON(w, 200, snippet)
  }
  ```

---

## Writing Robust Code

### Validate at the Boundary, Trust Internally

Validate user input (HTTP request bodies, query parameters) at the handler
layer. Do not add defensive checks inside `internal/db` for things that are
impossible given the callers you control.

### Complete if-else Chains

Every `if`/`else if` chain that handles an enum-like value must have a terminal
`else` that either panics with a descriptive message or returns an error.

- _Bad_
  ```go
  if mode == "fts" {
      // ...
  } else if mode == "tag" {
      // ...
  }
  // falls through silently if mode is something unexpected
  ```
- _Good_
  ```go
  if mode == "fts" {
      // ...
  } else if mode == "tag" {
      // ...
  } else {
      return fmt.Errorf("unknown list mode %q", mode)
  }
  ```

### Do Not Hardwire Magic Values

Constants and config values belong in `internal/config` or as named constants,
never buried in logic.

- _Bad_
  ```go
  if len(req.Password) < 8 {
  ```
- _Good_
  ```go
  const minPasswordLen = 8
  if len(req.Password) < minPasswordLen {
  ```

### Use `defer` for Cleanup

Always `defer rows.Close()` immediately after a successful `db.Query`. Always
`defer db.Close()` after `db.Open`. This prevents leaks even if the function
returns early.

```go
rows, err := db.Query(...)
if err != nil {
    return nil, err
}
defer rows.Close()
```

### Check `rows.Err()` After Iteration

A partial result is worse than no result. Always check the iterator error.

```go
for rows.Next() {
    // scan...
}
return snippets, rows.Err()
```

### No `init()` Functions

`init()` runs implicitly and is impossible to test in isolation. Put
initialization logic in explicit `Setup` or `Open` functions called from
`main`.

### No Panics Outside `main`

Library code (everything under `internal/`) must never `panic`. Use error
returns. Panics are reserved for truly unrecoverable situations, and only
`main` should call `log.Fatal`.

---

## Packages and Architecture

### Package Dependency Rules

```
models      → (nothing)
config      → (nothing)
db          → models
api         → db, models, config
cmd/server  → api, db, config
cmd/cli     → (TBD in Phase 2)
```

Violations are compilation errors if they create cycles; violations that don't
create cycles are still forbidden by this guide.

### Package Cohesion

Each package owns exactly one concept:

- `db` — all SQL, no HTTP, no business logic
- `api` — all HTTP handling, no SQL
- `models` — plain structs, no methods that touch the DB or HTTP
- `config` — env loading and struct definition

### File Names

File names are lowercase, single-word where possible. For `internal/api`, split
by resource: `auth.go`, `snippets.go`, `middleware.go`. For `internal/db`,
split by entity: `users.go`, `snippets.go`.

Do not create files like `utils.go` or `helpers.go` — these become catch-alls.
If something needs to be shared, ask which concept it belongs to.

---

## Logging

### Always Use `slog`, Never `fmt.Println`

```go
// bad
fmt.Println("server started on", port)

// good
slog.Info("snippr listening", "port", port)
```

### Log Levels

| Level | Use |
|---|---|
| `slog.Debug` | Internal state useful only while debugging a specific issue. Off by default. |
| `slog.Info` | Normal lifecycle events: server started, DB opened, migration ran. |
| `slog.Warn` | Unexpected but recoverable: a request with an unusual but valid payload. |
| `slog.Error` | Failures that the system couldn't handle: DB errors, encode failures. |

### Structured Logging — Key-Value Pairs

Always pass context as key-value pairs, not as format strings.

- _Bad_
  ```go
  slog.Error(fmt.Sprintf("encode response failed: %v", err))
  ```
- _Good_
  ```go
  slog.Error("encode response", "err", err)
  ```

### Log at the Site of Handling, Not Propagation

Log the error where you handle it (API layer). Do not log in `internal/db` —
just return the wrapped error. Logging while also returning causes double-log.

- _Bad_
  ```go
  // in db/snippets.go
  slog.Error("create snippet", "err", err)
  return 0, fmt.Errorf("create snippet: %w", err)
  ```
- _Good_
  ```go
  // in db/snippets.go — just return
  return 0, fmt.Errorf("create snippet: %w", err)

  // in api/snippets.go — log and respond
  if err != nil {
      slog.Error("create snippet", "err", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
  }
  ```

---

## Testing

### All DB Tests Use In-Memory SQLite

Never test against a file-backed DB. Use `:memory:` so tests are isolated and
fast.

```go
func setupDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := db.Open(":memory:")
    if err != nil {
        t.Fatal(err)
    }
    if err := db.RunMigrations(db); err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

### No Mocking the DB

Use real SQLite in tests. Mocks lie. We were bitten by mock/prod divergence;
real queries catch schema bugs and constraint violations.

### `t.Helper()` in Every Helper

Every function called from a test that calls `t.Fatal` / `t.Error` must call
`t.Helper()` first. Otherwise failure lines point to the helper, not the test.

```go
func post(t *testing.T, r http.Handler, path, body string) *httptest.ResponseRecorder {
    t.Helper()
    // ...
}
```

### Table-Driven Tests for Input Variation

When testing the same function with different inputs, use a table.

```go
tests := []struct {
    name     string
    email    string
    password string
    wantCode int
}{
    {"valid", "a@b.com", "password123", 201},
    {"short password", "a@b.com", "short", 400},
    {"invalid email", "notanemail", "password123", 400},
}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        // ...
    })
}
```

### Test the Behaviour, Not the Implementation

Do not call unexported functions in tests. Test via the exported API or the
HTTP layer.

### Test File Layout

- Test files live next to the code: `api/auth_test.go` tests `api/auth.go`.
- Shared test helpers go in `api/testhelper_test.go`.
- Package suffix is `_test` (black-box): `package api_test`.

---

## SQL

### Always List Columns Explicitly

Never `SELECT *`. Column order in `Scan` must match the query; explicit lists
make this verifiable by inspection and survive column additions.

- _Bad_
  ```go
  `SELECT * FROM snippets WHERE id = ?`
  ```
- _Good_
  ```go
  `SELECT id, user_id, title, content, language, is_public,
          COALESCE(share_slug, ''), created_at, updated_at
   FROM snippets WHERE id = ?`
  ```

### COALESCE for Nullable Columns

Use `COALESCE` in the query, scan into a plain Go type. Never use
`sql.NullString`, `sql.NullInt64`, etc.

- _Bad_
  ```go
  var slug sql.NullString
  rows.Scan(..., &slug)
  s.ShareSlug = slug.String
  ```
- _Good_
  ```go
  rows.Scan(..., &s.ShareSlug)  // query uses COALESCE(share_slug, '')
  ```

### Every User-Data Query Checks Ownership

Any query that reads or mutates a snippet must include `AND user_id = ?`.
Omitting it is a horizontal privilege escalation bug.

```go
`SELECT ... FROM snippets WHERE id = ? AND user_id = ?`
```

### SQLite Booleans Are Integers

Store booleans as `INTEGER (0/1)`. Convert at the scan boundary, not in SQL.

```go
var isPublic int
rows.Scan(..., &isPublic, ...)
s.IsPublic = isPublic == 1
```

### Query Indentation

Multi-line queries use backtick strings. Align keywords and columns:

```go
rows, err := db.Query(
    `SELECT s.id, s.user_id, s.title, s.content, s.language,
            s.is_public, COALESCE(s.share_slug, ''),
            s.created_at, s.updated_at
     FROM snippets s
     JOIN tags t ON t.snippet_id = s.id
     WHERE t.name = ? AND s.user_id = ?
     ORDER BY s.updated_at DESC`,
    tag, userID,
)
```

### Use Parameterized Queries — Always

Never build SQL strings with `fmt.Sprintf` or string concatenation. Positional
`?` placeholders prevent SQL injection.

- _Bad_
  ```go
  db.Query("SELECT ... WHERE email = '" + email + "'")
  ```
- _Good_
  ```go
  db.Query(`SELECT ... WHERE email = ?`, email)
  ```

---

## HTTP Handlers

### Handler Pattern: Factory Function

```go
func CreateSnippetHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // ...
    }
}
```

### Handler Body Order

Every handler follows this exact sequence:

1. **Decode** — parse the request body or URL params
2. **Validate** — check for required fields, format, length
3. **Authorize** — confirm the caller owns the resource
4. **Act** — call the DB function
5. **Respond** — `writeJSON` or `writeError`

```go
func UpdateSnippetHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Decode URL param.
        id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
        if err != nil {
            writeError(w, http.StatusBadRequest, "invalid id")
            return
        }
        // 2. Decode body.
        var req models.UpdateSnippetReq
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            writeError(w, http.StatusBadRequest, "invalid request body")
            return
        }
        // 3. Authorize — fetch snippet the user owns.
        userID := GetUserID(r)
        existing, err := dbpkg.GetSnippet(db, id, userID)
        if err != nil {
            writeError(w, http.StatusInternalServerError, "server error")
            return
        }
        if existing == nil {
            writeError(w, http.StatusNotFound, "snippet not found")
            return
        }
        // 4. Act.
        existing.Title = req.Title
        if err := dbpkg.UpdateSnippet(db, existing); err != nil {
            writeError(w, http.StatusInternalServerError, "server error")
            return
        }
        // 5. Respond.
        writeJSON(w, http.StatusOK, existing)
    }
}
```

### Request Structs Are Unexported and Local

`createSnippetReq`, `loginReq` etc. live in the same file as their handler and
are unexported. Do not reuse the DB model struct for decoding input.

### Use Standard HTTP Status Codes

| Situation | Code |
|---|---|
| Success (created) | 201 Created |
| Success (read/update) | 200 OK |
| Success (delete) | 204 No Content |
| Bad input | 400 Bad Request |
| Not authenticated | 401 Unauthorized |
| Authenticated but not allowed | 403 Forbidden |
| Resource not found | 404 Not Found |
| Resource already exists | 409 Conflict |
| Server error | 500 Internal Server Error |

### JSON Response Consistency

- Always return `[]` not `null` for empty lists.
  ```go
  if snippets == nil {
      snippets = []models.Snippet{}
  }
  ```
- Error responses always use `{"error": "message"}`.
- Success responses for list endpoints always use a JSON array.

---

## Concurrency

### No Goroutines in Handlers (Phase 1)

HTTP handlers in Phase 1 are fully synchronous. Do not spawn goroutines inside
handlers. If you need background work (email, backups), use a dedicated
background worker started from `main`.

### Protect Shared State With a Mutex

If you introduce any package-level mutable state (which you should avoid), it
must be protected with `sync.Mutex` or `sync.RWMutex`.

### Avoid `sync.WaitGroup` Leaks

Always call `wg.Add(n)` before spawning goroutines, not inside them. Always
`defer wg.Done()` as the first statement in the goroutine body.

---

## Tooling

### Linter: `go vet` Is Required to Pass

Run `go vet ./...` before every commit. Fix all reported issues; do not suppress
them without a comment explaining why.

### Formatter: `gofmt`

All code must be `gofmt`-formatted. If your editor does not auto-format on
save, run `gofmt -w .` before committing. Do not mix formatting changes with
logic changes in the same commit.

### When to Disable a Lint

If a specific check must be suppressed, add an inline comment with the reason:

```go
_ = db.Close() //nolint:errcheck — best-effort close in deferred cleanup
```

Never suppress without a reason.

### Makefile Targets

| Target | Runs |
|---|---|
| `make run` | `go run ./cmd/server` |
| `make build` | `go build -o bin/snippr ./cmd/server` |
| `make test` | `go test ./...` |
| `make lint` | `go vet ./...` |
| `make docker-build` | `docker build -t snippr .` |
| `make docker-up` | `docker compose up -d` |

---

## References

- [Effective Go](https://go.dev/doc/effective_go) — the canonical Go style reference
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) — official supplement
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) — comprehensive community guide
- [100 Go Mistakes](https://100go.co) — deep-dive into common pitfalls
- [causify-ai Python Style Guide](https://github.com/causify-ai/helpers/blob/master/docs/code_guidelines/all.coding_style.how_to_guide.md) — the inspiration for this document
