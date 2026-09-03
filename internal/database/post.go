package database

import(
	"lions/internal/models"
	"database/sql"
)

//create post return postid
func CreatePost(db *sql.DB, userID int64, title string, content string, categoryID int64) (int64,error) {
	
}