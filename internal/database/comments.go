package database

import (
	"database/sql"
	_ "embed"

	"lions/internal/models"
)

//go:embed sql/list_post.sql
var listCommentsQuery string

//success: return id, fail return err
func CreateComment(db *sql.DB, userID int64 , postID int64, content string) (int64,error) {
	query:= "INSERT INTO comments (user_id, post_id, content) VALUES (?,?,?)"
	result,err:= db.Exec(query, userID,postID,content)
	if err != nil {
		return 0, err
	}
	CommentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return CommentID, nil

}

func ListComments(db *sql.DB, postID int64) ([]models.Comment, error) {
	allComments:=[]models.Comment{}

	rows, err:= db.Query(listCommentsQuery, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next(){
		var comment models.Comment
		if err = rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.Author,
			&comment.Likes,
			&comment.Dislikes,

		) ; err!= nil{
			return nil, err
		}
		allComments = append(allComments, comment)
	}
	if err := rows.Err(); err != nil { // Check that iteration did not stop because of an error.
		return nil, err
	}

	return allComments, nil
	
}

