package database

import (
	"database/sql"
	"lions/internal/models"
	"time"
)

// Create user after they registered
func CreateUser(db *sql.DB, email, username, passwordHash string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)`,
		email, username, passwordHash)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Get user by email
func UserByEmail(db *sql.DB, email string) (*models.User, error) {
	u := &models.User{}
	err := db.QueryRow(
		`SELECT id, email, username, password_hash, created_at FROM users WHERE email = ?`,
		email).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Create session after user logged in
func CreateSession(db *sql.DB, id string, userID int64, expires time.Time) error {
	// One active session per user keeps stale cookies from piling up.
	if _, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return err
	}
	_, err := db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		id, userID, expires)
	return err
}

// Delete session after user logged out
func DeleteSession(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

// Search user if its session is still valid 
func UserBySession(db *sql.DB, sessionID string) (*models.User, error) {
	u := &models.User{}
	err := db.QueryRow(`
		SELECT u.id, u.email, u.username, u.created_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.id = ? AND s.expires_at > CURRENT_TIMESTAMP`,
		sessionID).Scan(&u.ID, &u.Email, &u.Username, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
