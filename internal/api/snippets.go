package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	dbpkg "github.com/vedanshu/snippr/internal/db"
	"github.com/vedanshu/snippr/internal/models"
)

func CreateSnippetHandler(db *sql.DB, hub *Hub) http.HandlerFunc {
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

		if req.WorkspaceID != 0 {
			ok, err := dbpkg.IsMember(db, req.WorkspaceID, userID)
			if err != nil {
				slog.Error("check workspace membership", "err", err)
				writeError(w, http.StatusInternalServerError, "server error")
				return
			}
			if !ok {
				writeError(w, http.StatusForbidden, "not a workspace member")
				return
			}
		}

		s := &models.Snippet{
			UserID:      userID,
			Title:       req.Title,
			Content:     req.Content,
			Language:    req.Language,
			IsPublic:    req.IsPublic,
			Tags:        req.Tags,
			WorkspaceID: req.WorkspaceID,
		}
		if s.Language == "" {
			s.Language = "text"
		}

		id, err := dbpkg.CreateSnippet(db, s)
		if err != nil {
			slog.Error("create snippet", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		s.ID = id
		writeJSON(w, http.StatusCreated, s)

		if s.WorkspaceID != 0 {
			if event, err := MakeBoardEvent("snippet_added", s); err == nil {
				hub.Broadcast(s.WorkspaceID, event)
			}
		}
	}
}

func ListSnippetsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		q := r.URL.Query().Get("q")
		tag := r.URL.Query().Get("tag")
		workspaceID, _ := strconv.ParseInt(r.URL.Query().Get("workspace_id"), 10, 64)

		if workspaceID != 0 {
			ok, err := dbpkg.IsMember(db, workspaceID, userID)
			if err != nil {
				slog.Error("check workspace membership", "err", err)
				writeError(w, http.StatusInternalServerError, "server error")
				return
			}
			if !ok {
				writeError(w, http.StatusForbidden, "not a workspace member")
				return
			}
		}

		snippets, err := dbpkg.ListSnippets(db, userID, workspaceID, q, tag)
		if err != nil {
			slog.Error("list snippets", "err", err)
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

		// Try personal ownership first; fall back to workspace membership.
		s, err := dbpkg.GetSnippet(db, id, userID)
		if err != nil {
			slog.Error("get snippet", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		if s == nil {
			// Not owned by user — check if it belongs to a workspace the user is a member of.
			s, err = dbpkg.GetWorkspaceSnippetForUser(db, id, userID)
			if err != nil {
				slog.Error("get workspace snippet", "err", err)
				writeError(w, http.StatusInternalServerError, "server error")
				return
			}
		}

		if s == nil {
			writeError(w, http.StatusNotFound, "snippet not found")
			return
		}

		writeJSON(w, http.StatusOK, s)
	}
}

func UpdateSnippetHandler(db *sql.DB, hub *Hub) http.HandlerFunc {
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
			slog.Error("get snippet for update", "err", err)
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
			slog.Error("update snippet", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		writeJSON(w, http.StatusOK, existing)

		if existing.WorkspaceID != 0 {
			if event, err := MakeBoardEvent("snippet_updated", existing); err == nil {
				hub.Broadcast(existing.WorkspaceID, event)
			}
		}
	}
}

func DeleteSnippetHandler(db *sql.DB, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)

		// Fetch before deleting so we know the workspace_id for broadcasting.
		existing, err := dbpkg.GetSnippet(db, id, userID)
		if err != nil {
			slog.Error("get snippet for delete", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		if err := dbpkg.DeleteSnippet(db, id, userID); err != nil {
			slog.Error("delete snippet", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)

		if existing != nil && existing.WorkspaceID != 0 {
			hub.Broadcast(existing.WorkspaceID, MakeDeleteEvent(id))
		}
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
