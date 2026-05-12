package db

import (
	"database/sql"
	"fmt"

	"github.com/vedanshu/snippr/internal/models"
)

func CreateUser(db *sql.DB, email, passwordHash string) (int64, error) {
	var id int64
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

// GetUserEmail returns the email for a user ID, or "" if not found.
func GetUserEmail(db *sql.DB, userID int64) (string, error) {
	var email string
	err := db.QueryRow(`SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get user email: %w", err)
	}
	return email, nil
}

func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	u := &models.User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}
