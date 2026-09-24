package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	email         TEXT NOT NULL UNIQUE COLLATE NOCASE, -- case intensive, standard for email and username
	username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
	password_hash TEXT NOT NULL,
	created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	profile_image BLOB
);

CREATE TABLE IF NOT EXISTS sessions (
	id         TEXT PRIMARY KEY,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
	id   INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE COLLATE NOCASE,
	kind TEXT NOT NULL CHECK (kind IN ('genre', 'book', 'discussion', 'author'))
);

CREATE TABLE IF NOT EXISTS posts (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title      TEXT NOT NULL,
	content    TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS post_categories (
	post_id     INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
	PRIMARY KEY (post_id, category_id)
);

CREATE TABLE IF NOT EXISTS comments (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	post_id    INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	content    TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS post_reactions (
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
	value   INTEGER NOT NULL CHECK (value IN (1, -1)), -- like/dislike
	PRIMARY KEY (user_id, post_id)
);

CREATE TABLE IF NOT EXISTS comment_reactions (
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	comment_id INTEGER NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
	value      INTEGER NOT NULL CHECK (value IN (1, -1)),
	PRIMARY KEY (user_id, comment_id)
);

CREATE TABLE IF NOT EXISTS files (
    id           INTEGER PRIMARY KEY,
    comment_id   INTEGER NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    filename     TEXT NOT NULL,
    content_type TEXT NOT NULL,
    data         BLOB NOT NULL,
    created_at   TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CHECK (length(data) > 0 AND length(data) <= 10 * 1024 * 1024),
    CHECK (content_type IN (
        'image/jpeg',
        'image/png',
        'image/webp',
        'application/pdf'
    ))
);

CREATE TABLE IF NOT EXISTS password_resets (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
`

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "forum.db?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("err apply schema: %w", err)
	}

	return db, nil
}
