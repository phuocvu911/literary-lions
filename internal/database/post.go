package database

import (
	"database/sql"
	_ "embed"

	"lions/internal/models"
)

//go:embed sql/list_post.sql
var listPostsQuery string

//go:embed sql/get_post.sql
var getPostQuery string

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

// ListPosts returns all posts ordered from newest to oldest.
func ListPosts(db *sql.DB) ([]models.Post, error) {
	allPosts := []models.Post{}

	rows, err := db.Query(listPostsQuery)
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
