package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"

	"github.com/vedanshu/snippr/internal/api"
	"github.com/vedanshu/snippr/internal/config"
	"github.com/vedanshu/snippr/internal/db"
)

//go:embed all:ui
var uiFS embed.FS

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("open database", "err", err)
		return
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		slog.Error("run migrations", "err", err)
		return
	}

	ja := jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(api.Logger)

	// API routes.
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

	// SPA static files — all non-API requests fall through to index.html.
	static, _ := fs.Sub(uiFS, "ui")
	r.Handle("/*", spaHandler(http.FS(static)))

	slog.Info("snippr listening", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server error", "err", err)
	}
}

// spaHandler serves static files and falls back to index.html for client-side routing.
func spaHandler(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := fs.Open(r.URL.Path)
		if err != nil {
			// File not found — serve index.html for React Router to handle.
			r2 := *r
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, &r2)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}
