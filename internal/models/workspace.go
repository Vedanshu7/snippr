package models

import "time"

// Workspace is a shared space that groups members and their snippets.
type Workspace struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	IsPublic   bool      `json:"is_public"`
	CreatedAt  time.Time `json:"created_at"`
}

// WorkspaceMember pairs a user with a workspace they belong to.
type WorkspaceMember struct {
	WorkspaceID int64     `json:"workspace_id"`
	UserID      int64     `json:"user_id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type CreateWorkspaceReq struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type UpdateWorkspaceReq struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type JoinWorkspaceReq struct {
	InviteCode  string `json:"invite_code"`
	WorkspaceID int64  `json:"workspace_id"` // used when joining a public workspace directly
}
