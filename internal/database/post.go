package database

import(
//	"lions/internal/models"
	"database/sql"

	
)

//create post return postid. Failure return 0, err
func CreatePost(db *sql.DB, userID int64, title string, content string, categoryIDs []int64) (int64,error) {
	tx, err:= db.Begin()
	if err!= nil{
		return 0, err
	}
	defer tx.Rollback()


	//edit new post
	sqlCreatePost:= "INSERT INTO posts (user_id, title, content) VALUES (?,?,?)"
	result, err:= tx.Exec(sqlCreatePost, userID,title, content)
	if err!= nil{
		return 0, err
	}
	postID,err:= result.LastInsertId()
	if err!= nil{
		return 0, err
	}
	//add post cate relations
	for _, cate:= range categoryIDs{
		sqlPostCategory:= "INSERT INTO post_categories (post_id, category_id) VALUES (?,?)"
		_,err := tx.Exec(sqlPostCategory, postID, cate)
		if err!= nil{
			return 0, err
		}
	}

	err=tx.Commit()
	if err!= nil{
		return 0, err
	}
	return postID,nil
}