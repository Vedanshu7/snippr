package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	dbpkg "github.com/vedanshu/snippr/internal/db"
	"github.com/vedanshu/snippr/internal/models"
)

func CreateSnippetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateSnippetReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(req.Title) == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}

		userID := GetUserID(r)
		s := &models.Snippet{
			UserID:   userID,
			Title:    req.Title,
			Content:  req.Content,
			Language: req.Language,
			IsPublic: req.IsPublic,
			Tags:     req.Tags,
		}
		if s.Language == "" {
			s.Language = "text"
		}

		id, err := dbpkg.CreateSnippet(db, s)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		s.ID = id
		writeJSON(w, http.StatusCreated, s)
	}
}

func ListSnippetsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		q := r.URL.Query().Get("q")
		tag := r.URL.Query().Get("tag")

		snippets, err := dbpkg.ListSnippets(db, userID, q, tag)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		if snippets == nil {
			snippets = []models.Snippet{}
		}
		writeJSON(w, http.StatusOK, snippets)
	}
}

func GetSnippetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)
		s, err := dbpkg.GetSnippet(db, id, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if s == nil {
			writeError(w, http.StatusNotFound, "snippet not found")
			return
		}

		writeJSON(w, http.StatusOK, s)
	}
}

func UpdateSnippetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req models.UpdateSnippetReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

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

		existing.Title = req.Title
		existing.Content = req.Content
		existing.Language = req.Language
		existing.IsPublic = req.IsPublic
		existing.Tags = req.Tags

		if err := dbpkg.UpdateSnippet(db, existing); err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		writeJSON(w, http.StatusOK, existing)
	}
}

func DeleteSnippetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)
		if err := dbpkg.DeleteSnippet(db, id, userID); err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func PublicSnippetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		s, err := dbpkg.GetSnippetBySlug(db, slug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if s == nil {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		writeJSON(w, http.StatusOK, s)
	}
}
