package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"lions/internal/database"
	"lions/internal/models"
	"log"
	"movies-api/models"
	"movies-api/service"
	"net/http"
	"strconv"
)


func (app *App) CreateComment(w http.ResponseWriter, r *http.Request, user *models.User) {
	var comment models.Comment
	if err := decodeJSON(r, &comment); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	postID, err:=strconv.Atoi(r.PathValue("id")) 
	if err!= nil{
		writeError(w, err, http.StatusBadRequest)
		return
	}
	
	commentID, err := database.CreateComment(app.db, user.ID, int64(postID), comment.Content)
	if err!= nil{
		writeError(w,err,http.StatusBadRequest)
		return
	}
	comment.ID = commentID

	writeJSON(w, http.StatusCreated, comment)
}

func (app *App)ListComment(w http.ResponseWriter, r *http.Request) {

}

