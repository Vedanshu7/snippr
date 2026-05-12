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

func CreateWorkspaceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateWorkspaceReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		userID := GetUserID(r)
		workspace, err := dbpkg.CreateWorkspace(db, userID, req.Name, req.IsPublic)
		if err != nil {
			slog.Error("create workspace", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		writeJSON(w, http.StatusCreated, workspace)
	}
}

func ListWorkspacesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		workspaces, err := dbpkg.ListWorkspaces(db, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		if workspaces == nil {
			workspaces = []models.Workspace{}
		}
		writeJSON(w, http.StatusOK, workspaces)
	}
}

func ListPublicWorkspacesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaces, err := dbpkg.ListPublicWorkspaces(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if workspaces == nil {
			workspaces = []models.Workspace{}
		}
		writeJSON(w, http.StatusOK, workspaces)
	}
}

func GetWorkspaceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)
		workspace, err := dbpkg.GetWorkspace(db, id, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if workspace == nil {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}

		writeJSON(w, http.StatusOK, workspace)
	}
}

func UpdateWorkspaceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req models.UpdateWorkspaceReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		userID := GetUserID(r)
		ok, err := dbpkg.IsAdmin(db, id, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if !ok {
			writeError(w, http.StatusForbidden, "only workspace admins can edit it")
			return
		}

		if err := dbpkg.UpdateWorkspace(db, id, userID, req.Name, req.IsPublic); err != nil {
			slog.Error("update workspace", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		workspace, err := dbpkg.GetWorkspace(db, id, userID)
		if err != nil || workspace == nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		writeJSON(w, http.StatusOK, workspace)
	}
}

func DeleteWorkspaceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)
		if err := dbpkg.DeleteWorkspace(db, id, userID); err != nil {
			if strings.Contains(err.Error(), "not owner") {
				writeError(w, http.StatusForbidden, "only the workspace admin can delete it")
				return
			}
			slog.Error("delete workspace", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func JoinWorkspaceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.JoinWorkspaceReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		userID := GetUserID(r)

		var workspace *models.Workspace
		var err error

		if req.WorkspaceID != 0 {
			// Direct join for public workspaces — verify it is actually public.
			workspace, err = dbpkg.GetPublicWorkspace(db, req.WorkspaceID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server error")
				return
			}
			if workspace == nil {
				writeError(w, http.StatusNotFound, "workspace not found or not public")
				return
			}
		} else {
			// Invite-code join for private workspaces.
			req.InviteCode = strings.TrimSpace(req.InviteCode)
			if req.InviteCode == "" {
				writeError(w, http.StatusBadRequest, "invite_code is required for private workspaces")
				return
			}
			workspace, err = dbpkg.GetWorkspaceByInviteCode(db, req.InviteCode)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server error")
				return
			}
			if workspace == nil {
				writeError(w, http.StatusNotFound, "invalid invite code")
				return
			}
		}

		if err := dbpkg.JoinWorkspace(db, workspace.ID, userID); err != nil {
			slog.Error("join workspace", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		writeJSON(w, http.StatusOK, workspace)
	}
}

func LeaveWorkspaceHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)
		if err := dbpkg.LeaveWorkspace(db, id, userID); err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ListWorkspaceMembersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)
		ok, err := dbpkg.IsMember(db, id, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if !ok {
			writeError(w, http.StatusForbidden, "not a workspace member")
			return
		}

		members, err := dbpkg.ListWorkspaceMembers(db, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if members == nil {
			members = []models.WorkspaceMember{}
		}
		writeJSON(w, http.StatusOK, members)
	}
}

func RemoveMemberHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid workspace id")
			return
		}
		targetID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		callerID := GetUserID(r)
		if err := dbpkg.RemoveMember(db, id, targetID, callerID); err != nil {
			msg := err.Error()
			switch {
			case strings.Contains(msg, "not admin"):
				writeError(w, http.StatusForbidden, "only workspace admins can remove members")
			case strings.Contains(msg, "cannot remove"):
				writeError(w, http.StatusForbidden, "cannot remove the workspace owner")
			default:
				slog.Error("remove member", "err", err)
				writeError(w, http.StatusInternalServerError, "server error")
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func RotateInviteCodeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		userID := GetUserID(r)

		// Require admin role (consistent with other admin-only actions).
		ok, err := dbpkg.IsAdmin(db, id, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if !ok {
			writeError(w, http.StatusForbidden, "only workspace admins can rotate the invite code")
			return
		}

		newCode, err := dbpkg.RotateInviteCode(db, id, userID)
		if err != nil {
			slog.Error("rotate invite code", "err", err)
			writeError(w, http.StatusInternalServerError, "server error")
			return
		}
		if newCode == "" {
			writeError(w, http.StatusForbidden, "only workspace admins can rotate the invite code")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"invite_code": newCode})
	}
}
