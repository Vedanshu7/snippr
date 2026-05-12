package api_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/vedanshu/snippr/internal/api"
	"github.com/vedanshu/snippr/internal/db"
)

const testSecret = "test-secret"

func testDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://snippr:snippr@localhost:5432/snippr_test?sslmode=disable"
	}
	return dsn
}

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(testDSN(t))
	if err != nil {
		t.Skipf("skipping: no test database available (%v)", err)
	}
	if err := db.RunMigrations(database); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Clean up test data between runs.
		database.Exec(`TRUNCATE users, snippets, tags RESTART IDENTITY CASCADE`)
		database.Close()
	})
	return database
}

func setupRouter(t *testing.T, database *sql.DB) (*chi.Mux, *jwtauth.JWTAuth) {
	t.Helper()
	ja := jwtauth.New("HS256", []byte(testSecret), nil)
	r := chi.NewRouter()

	r.Post("/api/register", api.RegisterHandler(database))
	r.Post("/api/login", api.LoginHandler(database, ja))
	r.Get("/api/s/{slug}", api.PublicSnippetHandler(database))

	r.Group(func(r chi.Router) {
		r.Use(api.Authenticator(ja))
		r.Post("/api/snippets", api.CreateSnippetHandler(database))
		r.Get("/api/snippets", api.ListSnippetsHandler(database))
		r.Get("/api/snippets/{id}", api.GetSnippetHandler(database))
		r.Put("/api/snippets/{id}", api.UpdateSnippetHandler(database))
		r.Delete("/api/snippets/{id}", api.DeleteSnippetHandler(database))
	})

	return r, ja
}
