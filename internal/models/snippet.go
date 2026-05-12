package models

import "time"

type Snippet struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	WorkspaceID int64     `json:"workspace_id,omitempty"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Language    string    `json:"language"`
	IsPublic    bool      `json:"is_public"`
	ShareSlug   string    `json:"share_slug,omitempty"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
	ID        int64  `json:"id"`
	SnippetID int64  `json:"snippet_id"`
	Name      string `json:"name"`
}

type CreateSnippetReq struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Language    string   `json:"language"`
	IsPublic    bool     `json:"is_public"`
	Tags        []string `json:"tags"`
	WorkspaceID int64    `json:"workspace_id,omitempty"`
}

type UpdateSnippetReq struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Language string   `json:"language"`
	IsPublic bool     `json:"is_public"`
	Tags     []string `json:"tags"`
}
