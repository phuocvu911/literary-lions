package database

import (
	"database/sql"
	"fmt"
	"time"
)

func Seed(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback()

	// ------------------------------------------------------------
	// Categories
	// ------------------------------------------------------------

	categories := []string{
		"Technology",
		"Programming",
		"Gaming",
		"Science",
		"News",
		"Music",
		"Movies",
		"Books",
		"Travel",
		"General",
	}

	categoryIDs := make(map[string]int64)

	for _, name := range categories {
		var id int64

		err := tx.QueryRow(`
			INSERT INTO categories (name)
			VALUES (?)
			ON CONFLICT(name) DO UPDATE SET name = excluded.name
			RETURNING id
		`, name).Scan(&id)

		if err != nil {
			return fmt.Errorf("insert category %q: %w", name, err)
		}

		categoryIDs[name] = id
	}

	// ------------------------------------------------------------
	// Users
	// ------------------------------------------------------------

	type User struct {
		ID       int64
		Username string
	}

	users := make([]User, 0, 10)

	for i := 1; i <= 10; i++ {
		username := fmt.Sprintf("user%d", i)
		email := fmt.Sprintf("user%d@example.com", i)

		var id int64

		err := tx.QueryRow(`
			INSERT INTO users (
				email,
				username,
				password_hash,
				created_at
			)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(username) DO UPDATE SET username = excluded.username
			RETURNING id
		`,
			email,
			username,
			"$2a$10$examplehashedpassword",
			time.Now().Add(-time.Duration(i)*24*time.Hour),
		).Scan(&id)

		if err != nil {
			return fmt.Errorf("insert user %q: %w", username, err)
		}

		users = append(users, User{
			ID:       id,
			Username: username,
		})
	}

	// ------------------------------------------------------------
	// Posts
	// ------------------------------------------------------------

	type Post struct {
		ID     int64
		UserID int64
	}

	posts := make([]Post, 0, 50)

	postCategories := []string{
		"Technology",
		"Programming",
		"Gaming",
		"Science",
		"News",
	}

	for _, user := range users {
		for postNum := 1; postNum <= 5; postNum++ {
			title := fmt.Sprintf(
				"Post %d by %s",
				postNum,
				user.Username,
			)

			content := fmt.Sprintf(
				"This is sample post %d created by %s. "+
					"It contains some example content for development and testing.",
				postNum,
				user.Username,
			)

			createdAt := time.Now().Add(
				-time.Duration(len(posts)+1) * time.Hour,
			)

			result, err := tx.Exec(`
				INSERT INTO posts (
					user_id,
					title,
					content,
					created_at
				)
				VALUES (?, ?, ?, ?)
			`,
				user.ID,
				title,
				content,
				createdAt,
			)

			if err != nil {
				return fmt.Errorf(
					"insert post for user %s: %w",
					user.Username,
					err,
				)
			}

			postID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("get post ID: %w", err)
			}

			posts = append(posts, Post{
				ID:     postID,
				UserID: user.ID,
			})

			// Give each post 1-2 categories.
			categoryName := postCategories[(postNum-1)%len(postCategories)]
			categoryID := categoryIDs[categoryName]

			_, err = tx.Exec(`
				INSERT OR IGNORE INTO post_categories (
					post_id,
					category_id
				)
				VALUES (?, ?)
			`, postID, categoryID)

			if err != nil {
				return fmt.Errorf("insert post category: %w", err)
			}

			// Add a second category to every second post.
			if postNum%2 == 0 {
				secondCategory := categories[(postNum+len(posts))%len(categories)]
				secondCategoryID := categoryIDs[secondCategory]

				_, err = tx.Exec(`
					INSERT OR IGNORE INTO post_categories (
						post_id,
						category_id
					)
					VALUES (?, ?)
				`, postID, secondCategoryID)

				if err != nil {
					return fmt.Errorf("insert second post category: %w", err)
				}
			}
		}
	}

	// ------------------------------------------------------------
	// Comments
	// ------------------------------------------------------------

	type Comment struct {
		ID     int64
		UserID int64
		PostID int64
	}

	comments := make([]Comment, 0, len(posts)*5)

	for postIndex, post := range posts {
		for commentNum := 1; commentNum <= 5; commentNum++ {
			// Pick a user different from the post author when possible.
			commentUser := users[
				(postIndex+commentNum)%len(users),
			]

			content := fmt.Sprintf(
				"Sample comment %d on post %d by %s.",
				commentNum,
				post.ID,
				commentUser.Username,
			)

			createdAt := time.Now().Add(
				-time.Duration(postIndex*5+commentNum) * time.Minute,
			)

			result, err := tx.Exec(`
				INSERT INTO comments (
					post_id,
					user_id,
					content,
					created_at
				)
				VALUES (?, ?, ?, ?)
			`,
				post.ID,
				commentUser.ID,
				content,
				createdAt,
			)

			if err != nil {
				return fmt.Errorf("insert comment: %w", err)
			}

			commentID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("get comment ID: %w", err)
			}

			comments = append(comments, Comment{
				ID:     commentID,
				UserID: commentUser.ID,
				PostID: post.ID,
			})
		}
	}

	// ------------------------------------------------------------
	// Post reactions
	// ------------------------------------------------------------

	for postIndex, post := range posts {
		// Give several users reactions to each post.
		for i := 1; i <= 5; i++ {
			user := users[(postIndex+i)%len(users)]

			value := 1
			if (postIndex+i)%4 == 0 {
				value = -1
			}

			_, err := tx.Exec(`
				INSERT OR IGNORE INTO post_reactions (
					user_id,
					post_id,
					value
				)
				VALUES (?, ?, ?)
			`,
				user.ID,
				post.ID,
				value,
			)

			if err != nil {
				return fmt.Errorf("insert post reaction: %w", err)
			}
		}
	}

	// ------------------------------------------------------------
	// Comment reactions
	// ------------------------------------------------------------

	for commentIndex, comment := range comments {
		// Give 3 users reactions to each comment.
		for i := 1; i <= 3; i++ {
			user := users[(commentIndex+i)%len(users)]

			value := 1
			if (commentIndex+i)%5 == 0 {
				value = -1
			}

			_, err := tx.Exec(`
				INSERT OR IGNORE INTO comment_reactions (
					user_id,
					comment_id,
					value
				)
				VALUES (?, ?, ?)
			`,
				user.ID,
				comment.ID,
				value,
			)

			if err != nil {
				return fmt.Errorf("insert comment reaction: %w", err)
			}
		}
	}

	// ------------------------------------------------------------
	// Files
	// ------------------------------------------------------------

	for i, user := range users {
		filename := fmt.Sprintf("avatar-%d.txt", i+1)
		content := []byte(fmt.Sprintf(
			"Sample file belonging to %s",
			user.Username,
		))

		_, err := tx.Exec(`
			INSERT INTO files (
				user_id,
				filename,
				content_type,
				size,
				data,
				created_at
			)
			VALUES (?, ?, ?, ?, ?, ?)
		`,
			user.ID,
			filename,
			"text/plain",
			len(content),
			content,
			time.Now().Add(-time.Duration(i)*24*time.Hour),
		)

		if err != nil {
			return fmt.Errorf("insert file: %w", err)
		}
	}

	// ------------------------------------------------------------
	// Commit
	// ------------------------------------------------------------

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	return nil
}
