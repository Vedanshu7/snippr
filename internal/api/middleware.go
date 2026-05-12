package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

// Authenticator verifies JWTs from either the Authorization header or the
// session cookie, so both the web UI (cookie) and CLI (Bearer token) work.
func Authenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		chain := jwtauth.Verifier(ja)(jwtauth.Authenticator(ja)(next))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Inject session cookie into Authorization header before the verifier runs.
			if r.Header.Get("Authorization") == "" {
				if c, err := r.Cookie("session"); err == nil {
					r = r.Clone(r.Context())
					r.Header.Set("Authorization", "Bearer "+c.Value)
				}
			}
			chain.ServeHTTP(w, r)
		})
	}
}

func GetUserID(r *http.Request) int64 {
	if id, ok := r.Context().Value(userIDKey).(int64); ok {
		return id
	}
	_, claims, _ := jwtauth.FromContext(r.Context())
	if id, ok := claims["user_id"].(float64); ok {
		return int64(id)
	}
	return 0
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
