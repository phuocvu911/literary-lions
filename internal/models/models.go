package models

import "time"

// User represents a user account in the forum, has exactly same fields as the table user in the database
type User struct {
	ID           int64
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// Category represents a category of posts, has exactly same fields as the table category in the database
type Category struct {
	ID   int64
	Name string
}

// Post represents a post in the forum.
type Post struct {
	ID            int64
	Title         string
	Content       string
	CreatedAt     time.Time
	Author        string
	Likes         int
	Dislikes      int
	CommentCount  int
	CategoryNames string
}

// Comment represents a comment on a post.
type Comment struct {
	ID        int64
	Content   string
	CreatedAt time.Time
	Author    string
	Likes     int
	Dislikes  int
}
