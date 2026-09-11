package database

import (
	"database/sql"
	_ "embed"
	"fmt"

	"lions/internal/models"
)

//go:embed sql/list_post.sql
var listPostsQuery string

//go:embed sql/get_post.sql
var getPostQuery string

//go:embed sql/search_posts.sql
var searchQuery string

// create post return postid. Failure return 0, err
func CreatePost(db *sql.DB, userID int64, title string, content string, categoryIDs []int64) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	//edit new post
	sqlCreatePost := "INSERT INTO posts (user_id, title, content) VALUES (?,?,?)"
	result, err := tx.Exec(sqlCreatePost, userID, title, content)
	if err != nil {
		return 0, err
	}
	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	//add post cate relations
	for _, cate := range categoryIDs {
		sqlPostCategory := "INSERT INTO post_categories (post_id, category_id) VALUES (?,?)"
		_, err := tx.Exec(sqlPostCategory, postID, cate)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return postID, nil
}

// ListPosts returns one page of posts ordered from newest to oldest.
func ListPosts(db *sql.DB, limit, offset int) ([]models.Post, error) {
	if limit <= 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination values")
	}

	allPosts := []models.Post{}

	rows, err := db.Query(listPostsQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
		if err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Author,
			&post.CommentCount,
			&post.Likes,
			&post.Dislikes,
			&post.CategoryNames,
		); err != nil {
			return nil, err
		}
		allPosts = append(allPosts, post)
	}
	if err := rows.Err(); err != nil { // Check that iteration did not stop because of an error.
		return nil, err
	}

	return allPosts, nil
}

// CountPosts returns the total number of posts before pagination.
func CountPosts(db *sql.DB) (int, error) {
	var total int
	err := db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func GetPostByID(db *sql.DB, id int64) (models.Post, error) {
	post := models.Post{}
	row := db.QueryRow(getPostQuery, id)
	if err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.Author,
		&post.CommentCount,
		&post.Likes,
		&post.Dislikes,
		&post.CategoryNames,
	); err != nil {
		return models.Post{}, err
	}

	return post, nil
}

func SearchPost(db *sql.DB, keyWord string, limit, offset int) ([]models.Post, error) {
	if limit <= 0 || offset < 0 {
		return nil, fmt.Errorf("invalid pagination values")
	}

	searchResults := []models.Post{}
	searchPattern := "%" + keyWord + "%"
	rows, err := db.Query(searchQuery, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
		if err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Author,
			&post.CommentCount,
			&post.Likes,
			&post.Dislikes,
			&post.CategoryNames,
		); err != nil {
			return nil, err
		}
		searchResults = append(searchResults, post)
	}
	if err := rows.Err(); err != nil { // Check that iteration did not stop because of an error.
		return nil, err
	}

	return searchResults, nil
}

// CountSearchPosts returns the number of posts matching a keyword before pagination.
func CountSearchPosts(db *sql.DB, keyWord string) (int, error) {
	searchPattern := "%" + keyWord + "%"

	var total int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM posts
		WHERE content LIKE ? OR title LIKE ?
	`, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
