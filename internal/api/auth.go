package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"

	dbpkg "github.com/vedanshu/snippr/internal/db"
)

var (
	reUpper   = regexp.MustCompile(`[A-Z]`)
	reDigit   = regexp.MustCompile(`[0-9]`)
	reSpecial = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func validatePassword(p string) string {
	switch {
	case len(p) < minPasswordLen:
		return "password must be at least 8 characters"
	case !reUpper.MatchString(p):
		return "password must contain at least one uppercase letter"
	case !reDigit.MatchString(p):
		return "password must contain at least one number"
	case !reSpecial.MatchString(p):
		return "password must contain at least one special character"
	}
	return ""
}

const minPasswordLen = 8

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || !strings.Contains(req.Email, "@") {
			writeError(w, http.StatusBadRequest, "invalid email")
			return
		}
		if msg := validatePassword(req.Password); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		existing, err := dbpkg.GetUserByEmail(db, req.Email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if existing != nil {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		id, err := dbpkg.CreateUser(db, req.Email, string(hash))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"id": id, "email": req.Email})
	}
}

func LoginHandler(db *sql.DB, ja *jwtauth.JWTAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		user, err := dbpkg.GetUserByEmail(db, req.Email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if user == nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		claims := map[string]any{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}
		_, tokenStr, err := ja.Encode(claims)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		// Set httpOnly cookie for web clients.
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    tokenStr,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   os.Getenv("APP_ENV") == "production",
			Path:     "/",
			MaxAge:   86400,
		})

		// Also return token in body for CLI compatibility.
		writeJSON(w, http.StatusOK, map[string]any{
			"token": tokenStr,
			"user":  map[string]any{"id": user.ID, "email": user.Email},
		})
	}
}

func LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   "",
			Path:    "/",
			MaxAge:  -1,
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusNoContent)
	}
}

func MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		id, _ := claims["user_id"].(float64)
		email, _ := claims["email"].(string)
		writeJSON(w, http.StatusOK, map[string]any{
			"id":    int64(id),
			"email": email,
		})
	}
}
