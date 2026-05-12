package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/vedanshu/snippr/internal/models"
)

func generateInviteCode() (string, error) {
	return generateSlug() // reuses the same 8-char base-62 generator
}

// CreateWorkspace creates a new workspace owned by ownerID and adds them as admin.
func CreateWorkspace(db *sql.DB, ownerID int64, name string, isPublic bool) (*models.Workspace, error) {
	var code string
	var err error

	// Retry up to 3 times in case of invite_code collision (astronomically unlikely).
	for range 3 {
		code, err = generateInviteCode()
		if err != nil {
			return nil, fmt.Errorf("generate invite code: %w", err)
		}

		w := &models.Workspace{}
		err = db.QueryRow(
			`INSERT INTO workspaces (owner_id, name, invite_code, is_public)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, owner_id, name, invite_code, is_public, created_at`,
			ownerID, name, code, isPublic,
		).Scan(&w.ID, &w.OwnerID, &w.Name, &w.InviteCode, &w.IsPublic, &w.CreatedAt)
		if err == nil {
			// Add owner as first member with admin role.
			if _, err2 := db.Exec(
				`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'admin')`,
				w.ID, ownerID,
			); err2 != nil {
				return nil, fmt.Errorf("add owner as member: %w", err2)
			}
			return w, nil
		}
		if !isUniqueViolation(err) {
			break
		}
	}
	return nil, fmt.Errorf("create workspace: %w", err)
}

// GetWorkspace returns the workspace only if userID is a member.
func GetWorkspace(db *sql.DB, id, userID int64) (*models.Workspace, error) {
	w := &models.Workspace{}
	err := db.QueryRow(
		`SELECT w.id, w.owner_id, w.name, w.invite_code, w.is_public, w.created_at
		 FROM workspaces w
		 JOIN workspace_members wm ON wm.workspace_id = w.id
		 WHERE w.id = $1 AND wm.user_id = $2`,
		id, userID,
	).Scan(&w.ID, &w.OwnerID, &w.Name, &w.InviteCode, &w.IsPublic, &w.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return w, nil
}

// ListWorkspaces returns all workspaces the user belongs to.
func ListWorkspaces(db *sql.DB, userID int64) ([]models.Workspace, error) {
	rows, err := db.Query(
		`SELECT w.id, w.owner_id, w.name, w.invite_code, w.is_public, w.created_at
		 FROM workspaces w
		 JOIN workspace_members wm ON wm.workspace_id = w.id
		 WHERE wm.user_id = $1
		 ORDER BY w.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()

	var out []models.Workspace
	for rows.Next() {
		var w models.Workspace
		if err := rows.Scan(&w.ID, &w.OwnerID, &w.Name, &w.InviteCode, &w.IsPublic, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ListPublicWorkspaces returns all publicly discoverable workspaces.
func ListPublicWorkspaces(db *sql.DB) ([]models.Workspace, error) {
	rows, err := db.Query(
		`SELECT id, owner_id, name, invite_code, is_public, created_at
		 FROM workspaces WHERE is_public = TRUE
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list public workspaces: %w", err)
	}
	defer rows.Close()

	var out []models.Workspace
	for rows.Next() {
		var w models.Workspace
		if err := rows.Scan(&w.ID, &w.OwnerID, &w.Name, &w.InviteCode, &w.IsPublic, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// GetWorkspaceByInviteCode looks up a workspace by its invite code.
func GetWorkspaceByInviteCode(db *sql.DB, code string) (*models.Workspace, error) {
	w := &models.Workspace{}
	err := db.QueryRow(
		`SELECT id, owner_id, name, invite_code, is_public, created_at
		 FROM workspaces WHERE invite_code = $1`,
		code,
	).Scan(&w.ID, &w.OwnerID, &w.Name, &w.InviteCode, &w.IsPublic, &w.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get workspace by invite code: %w", err)
	}
	return w, nil
}

// GetPublicWorkspace returns a workspace by ID only if it is public. Used for join-without-code.
func GetPublicWorkspace(db *sql.DB, id int64) (*models.Workspace, error) {
	w := &models.Workspace{}
	err := db.QueryRow(
		`SELECT id, owner_id, name, invite_code, is_public, created_at
		 FROM workspaces WHERE id = $1 AND is_public = TRUE`,
		id,
	).Scan(&w.ID, &w.OwnerID, &w.Name, &w.InviteCode, &w.IsPublic, &w.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get public workspace: %w", err)
	}
	return w, nil
}

// JoinWorkspace adds userID to the workspace as a member. Idempotent.
func JoinWorkspace(db *sql.DB, workspaceID, userID int64) error {
	_, err := db.Exec(
		`INSERT INTO workspace_members (workspace_id, user_id, role)
		 VALUES ($1, $2, 'member') ON CONFLICT DO NOTHING`,
		workspaceID, userID,
	)
	if err != nil {
		return fmt.Errorf("join workspace: %w", err)
	}
	return nil
}

// LeaveWorkspace removes userID from the workspace.
func LeaveWorkspace(db *sql.DB, workspaceID, userID int64) error {
	_, err := db.Exec(
		`DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	)
	if err != nil {
		return fmt.Errorf("leave workspace: %w", err)
	}
	return nil
}

// ListWorkspaceMembers returns all members of a workspace with email and role.
func ListWorkspaceMembers(db *sql.DB, workspaceID int64) ([]models.WorkspaceMember, error) {
	rows, err := db.Query(
		`SELECT wm.workspace_id, wm.user_id, u.email, wm.role, wm.joined_at
		 FROM workspace_members wm
		 JOIN users u ON u.id = wm.user_id
		 WHERE wm.workspace_id = $1
		 ORDER BY wm.joined_at ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("list workspace members: %w", err)
	}
	defer rows.Close()

	var out []models.WorkspaceMember
	for rows.Next() {
		var m models.WorkspaceMember
		if err := rows.Scan(&m.WorkspaceID, &m.UserID, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpdateWorkspace updates a workspace's name and visibility. Only the owner can do this.
func UpdateWorkspace(db *sql.DB, workspaceID, ownerID int64, name string, isPublic bool) error {
	result, err := db.Exec(
		`UPDATE workspaces SET name=$1, is_public=$2 WHERE id=$3 AND owner_id=$4`,
		name, isPublic, workspaceID, ownerID,
	)
	if err != nil {
		return fmt.Errorf("update workspace: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not owner")
	}
	return nil
}

// DeleteWorkspace deletes the workspace. Only the owner can do this.
// Snippets with this workspace_id will have workspace_id set to NULL (ON DELETE SET NULL).
func DeleteWorkspace(db *sql.DB, workspaceID, ownerID int64) error {
	result, err := db.Exec(
		`DELETE FROM workspaces WHERE id=$1 AND owner_id=$2`,
		workspaceID, ownerID,
	)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not owner")
	}
	return nil
}

// RemoveMember removes targetUserID from the workspace.
// The caller must be an admin, and the target cannot be the workspace owner.
func RemoveMember(db *sql.DB, workspaceID, targetUserID, adminUserID int64) error {
	ok, err := IsAdmin(db, workspaceID, adminUserID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("not admin")
	}

	result, err := db.Exec(
		`DELETE FROM workspace_members
		 WHERE workspace_id = $1 AND user_id = $2
		   AND $2 != (SELECT owner_id FROM workspaces WHERE id = $1)`,
		workspaceID, targetUserID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("cannot remove: target is owner or not a member")
	}
	return nil
}

// RotateInviteCode generates a new invite code for the workspace.
// Returns empty string if caller is not the owner.
func RotateInviteCode(db *sql.DB, workspaceID, ownerID int64) (string, error) {
	code, err := generateInviteCode()
	if err != nil {
		return "", fmt.Errorf("generate invite code: %w", err)
	}

	result, err := db.Exec(
		`UPDATE workspaces SET invite_code=$1 WHERE id=$2 AND owner_id=$3`,
		code, workspaceID, ownerID,
	)
	if err != nil {
		return "", fmt.Errorf("rotate invite code: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return "", nil // not the owner
	}
	return code, nil
}

// IsMember reports whether userID belongs to workspaceID.
func IsMember(db *sql.DB, workspaceID, userID int64) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`,
		workspaceID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("is member: %w", err)
	}
	return count > 0, nil
}

// IsAdmin reports whether userID is an admin of workspaceID.
func IsAdmin(db *sql.DB, workspaceID, userID int64) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM workspace_members WHERE workspace_id=$1 AND user_id=$2 AND role='admin'`,
		workspaceID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("is admin: %w", err)
	}
	return count > 0, nil
}

// isUniqueViolation detects a Postgres unique-constraint error by SQLSTATE.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505")
}
