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

	type seedCategory struct {
		Name string
		Kind string
	}

	categories := []seedCategory{
		{Name: "Fantasy", Kind: "genre"},
		{Name: "Science Fiction", Kind: "genre"},
		{Name: "Mystery", Kind: "genre"},
		{Name: "Classics", Kind: "genre"},
		{Name: "Non-fiction", Kind: "genre"},
		{Name: "Dune", Kind: "book"},
		{Name: "Harry Potter", Kind: "book"},
		{Name: "Book Reviews", Kind: "discussion"},
		{Name: "Theme Analysis", Kind: "discussion"},
		{Name: "Character Study", Kind: "discussion"},
		{Name: "J. K. Rowling", Kind: "author"},
		{Name: "Frank Herbert", Kind: "author"},
	}

	categoryIDs := make(map[string]int64)

	for _, category := range categories {
		var id int64

		err := tx.QueryRow(`
			INSERT INTO categories (name, kind)
			VALUES (?, ?)
			ON CONFLICT(name) DO UPDATE SET kind = excluded.kind
			RETURNING id
		`, category.Name, category.Kind).Scan(&id)

		if err != nil {
			return fmt.Errorf("insert category %q: %w", category.Name, err)
		}

		categoryIDs[category.Name] = id
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
				password_hash
			)
			VALUES (?, ?, ?)
			ON CONFLICT(username) DO UPDATE SET username = excluded.username
			RETURNING id
		`,
			email,
			username,
			"a4509dbd97b400cc759e42104f1c714f7a93ab15e50d7efa4108b441283d5a02",
		).Scan(&id)

		if err != nil {
			return fmt.Errorf("insert user %q: %w", username, err)
		}

		users = append(users, User{
			ID:       id,
			Username: username,
		})
	}

	_, _ = tx.Exec(`
			INSERT INTO users (
				email,
				username,
				password_hash
			)
			VALUES (?, ?, ?)`,
		"phuocvu@gmail.com",
		"hoang",
		"$2a$10$gGwFscqWfrx1rleRTINH1e7qUA.fPb5.v62Fy2EviEQtJcauDNRbq")

	// ------------------------------------------------------------
	// Posts
	// ------------------------------------------------------------

	type Post struct {
		ID     int64
		UserID int64
	}

	posts := make([]Post, 0, 50)

	postCategories := []string{
		"Fantasy",
		"Science Fiction",
		"Mystery",
		"Book Reviews",
		"Theme Analysis",
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
				secondCategory := categories[(postNum+len(posts))%len(categories)].Name
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
			commentUser := users[(postIndex+commentNum)%len(users)]

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
