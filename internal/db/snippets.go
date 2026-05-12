package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"

	"github.com/vedanshu/snippr/internal/models"
)

const slugChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateSlug() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = slugChars[int(b[i])%len(slugChars)]
	}
	return string(b), nil
}

func CreateSnippet(db *sql.DB, s *models.Snippet) (int64, error) {
	if s.IsPublic && s.ShareSlug == "" {
		slug, err := generateSlug()
		if err != nil {
			return 0, err
		}
		s.ShareSlug = slug
	}

	var id int64
	err := db.QueryRow(
		`INSERT INTO snippets (user_id, title, content, language, is_public, share_slug, workspace_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		s.UserID, s.Title, s.Content, s.Language, s.IsPublic,
		nullableSlug(s.ShareSlug), nullableID(s.WorkspaceID),
	).Scan(&id, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("create snippet: %w", err)
	}

	if err := setTags(db, id, s.Tags); err != nil {
		return 0, err
	}
	return id, nil
}

func GetSnippet(db *sql.DB, id, userID int64) (*models.Snippet, error) {
	s := &models.Snippet{}
	err := db.QueryRow(
		`SELECT id, user_id, title, content, language, is_public, COALESCE(share_slug,''),
		        COALESCE(workspace_id, 0), created_at, updated_at
		 FROM snippets WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Content, &s.Language, &s.IsPublic,
		&s.ShareSlug, &s.WorkspaceID, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get snippet: %w", err)
	}

	s.Tags, _ = getTags(db, s.ID)
	return s, nil
}

// GetWorkspaceSnippet returns a snippet by ID if it belongs to the given workspace.
func GetWorkspaceSnippet(db *sql.DB, id, workspaceID int64) (*models.Snippet, error) {
	s := &models.Snippet{}
	err := db.QueryRow(
		`SELECT id, user_id, title, content, language, is_public, COALESCE(share_slug,''),
		        COALESCE(workspace_id, 0), created_at, updated_at
		 FROM snippets WHERE id = $1 AND workspace_id = $2`,
		id, workspaceID,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Content, &s.Language, &s.IsPublic,
		&s.ShareSlug, &s.WorkspaceID, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get workspace snippet: %w", err)
	}

	s.Tags, _ = getTags(db, s.ID)
	return s, nil
}

// GetWorkspaceSnippetForUser returns a snippet by ID if it belongs to a workspace
// that userID is a member of. Used to allow non-author members to read workspace snippets.
func GetWorkspaceSnippetForUser(db *sql.DB, id, userID int64) (*models.Snippet, error) {
	s := &models.Snippet{}
	err := db.QueryRow(
		`SELECT s.id, s.user_id, s.title, s.content, s.language, s.is_public,
		        COALESCE(s.share_slug,''), COALESCE(s.workspace_id, 0), s.created_at, s.updated_at
		 FROM snippets s
		 JOIN workspace_members wm ON wm.workspace_id = s.workspace_id
		 WHERE s.id = $1 AND wm.user_id = $2`,
		id, userID,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Content, &s.Language, &s.IsPublic,
		&s.ShareSlug, &s.WorkspaceID, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get workspace snippet for user: %w", err)
	}

	s.Tags, _ = getTags(db, s.ID)
	return s, nil
}

func GetSnippetBySlug(db *sql.DB, slug string) (*models.Snippet, error) {
	s := &models.Snippet{}
	err := db.QueryRow(
		`SELECT id, user_id, title, content, language, is_public, share_slug,
		        COALESCE(workspace_id, 0), created_at, updated_at
		 FROM snippets WHERE share_slug = $1 AND is_public = TRUE`,
		slug,
	).Scan(&s.ID, &s.UserID, &s.Title, &s.Content, &s.Language, &s.IsPublic,
		&s.ShareSlug, &s.WorkspaceID, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get snippet by slug: %w", err)
	}

	s.Tags, _ = getTags(db, s.ID)
	return s, nil
}

func ListSnippets(db *sql.DB, userID, workspaceID int64, query, tag string) ([]models.Snippet, error) {
	var rows *sql.Rows
	var err error

	switch {
	case workspaceID != 0:
		rows, err = db.Query(
			`SELECT id, user_id, title, content, language, is_public,
			        COALESCE(share_slug,''), COALESCE(workspace_id,0), created_at, updated_at
			 FROM snippets WHERE workspace_id = $1
			 ORDER BY updated_at DESC`,
			workspaceID,
		)
	case query != "":
		rows, err = db.Query(
			`SELECT id, user_id, title, content, language, is_public,
			        COALESCE(share_slug,''), COALESCE(workspace_id,0), created_at, updated_at
			 FROM snippets
			 WHERE search_vector @@ plainto_tsquery('english', $1)
			   AND user_id = $2 AND workspace_id IS NULL
			 ORDER BY updated_at DESC`,
			query, userID,
		)
	case tag != "":
		rows, err = db.Query(
			`SELECT s.id, s.user_id, s.title, s.content, s.language, s.is_public,
			        COALESCE(s.share_slug,''), COALESCE(s.workspace_id,0), s.created_at, s.updated_at
			 FROM snippets s
			 JOIN tags t ON t.snippet_id = s.id
			 WHERE t.name = $1 AND s.user_id = $2 AND s.workspace_id IS NULL
			 ORDER BY s.updated_at DESC`,
			tag, userID,
		)
	default:
		rows, err = db.Query(
			`SELECT id, user_id, title, content, language, is_public,
			        COALESCE(share_slug,''), COALESCE(workspace_id,0), created_at, updated_at
			 FROM snippets WHERE user_id = $1 AND workspace_id IS NULL
			 ORDER BY updated_at DESC`,
			userID,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("list snippets: %w", err)
	}
	defer rows.Close()

	var snippets []models.Snippet
	for rows.Next() {
		var s models.Snippet
		if err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.Content, &s.Language,
			&s.IsPublic, &s.ShareSlug, &s.WorkspaceID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Tags, _ = getTags(db, s.ID)
		snippets = append(snippets, s)
	}
	return snippets, rows.Err()
}

func UpdateSnippet(db *sql.DB, s *models.Snippet) error {
	if s.IsPublic && s.ShareSlug == "" {
		slug, err := generateSlug()
		if err != nil {
			return err
		}
		s.ShareSlug = slug
	}

	_, err := db.Exec(
		`UPDATE snippets
		 SET title=$1, content=$2, language=$3, is_public=$4, share_slug=$5, updated_at=NOW()
		 WHERE id=$6 AND user_id=$7`,
		s.Title, s.Content, s.Language, s.IsPublic, nullableSlug(s.ShareSlug), s.ID, s.UserID,
	)
	if err != nil {
		return fmt.Errorf("update snippet: %w", err)
	}
	return setTags(db, s.ID, s.Tags)
}

func DeleteSnippet(db *sql.DB, id, userID int64) error {
	_, err := db.Exec(`DELETE FROM snippets WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func setTags(db *sql.DB, snippetID int64, tags []string) error {
	if _, err := db.Exec(`DELETE FROM tags WHERE snippet_id = $1`, snippetID); err != nil {
		return err
	}
	for _, name := range tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, err := db.Exec(
			`INSERT INTO tags (snippet_id, name) VALUES ($1, $2)`, snippetID, name,
		); err != nil {
			return err
		}
	}
	return nil
}

func getTags(db *sql.DB, snippetID int64) ([]string, error) {
	rows, err := db.Query(`SELECT name FROM tags WHERE snippet_id = $1 ORDER BY id`, snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, rows.Err()
}

func nullableSlug(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableID(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}
