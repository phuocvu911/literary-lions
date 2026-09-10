package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"lions/internal/database"
	"lions/internal/models"
	"strings"
	"net/http"
	"strconv"
)

type CommentListResponse struct{
	Items []models.Comment  `json:"items"`
	Page int `json: "page"`
	Limit int `json:"limit"`
	TotalItems int `json:"totalItems"`
}


func (app *App) CreateComment(w http.ResponseWriter, r *http.Request, user *models.User) {
	var comment models.Comment
	if err := decodeJSON(r, &comment); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	//check content is not empty
	if strings.TrimSpace(comment.Content) =="" {
		writeError(w, errors.New("Comments cannot be empty"), http.StatusBadRequest)
		return
	}
	postID:=GetPostIDFromUrl(w,r,"id")
	
	commentID, err := database.CreateComment(app.db, user.ID, int64(postID), comment.Content)
	if err!= nil{
		writeError(w,err,http.StatusBadRequest)
		return
	}
	comment.ID = commentID

	writeJSON(w, http.StatusCreated, comment)
}

func (app *App)ListComment(w http.ResponseWriter, r *http.Request) {
	limit,err:= optionalPageInt(r, "limit",20)
	if err!=nil{
		writeError(w, err, http.StatusBadRequest)
		return
	}
	page,err:= optionalPageInt(r, "page",0)
	if err!=nil{
		writeError(w, err, http.StatusBadRequest)
		return
	}
	offset:= limit*page
	postID:=GetPostIDFromUrl(w,r,"id")

	comments, err:= database.ListComments(app.db, int64(postID), limit, offset)
	if err!= nil{
		writeError(w,err,http.StatusBadRequest)
		return
	}

	totalItems:= len(comments)

	writeJSON(w,http.StatusOK, CommentListResponse{comments, page, limit, totalItems})



}

func optionalPageInt(r *http.Request, name string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(name + " must be an integer")
	}
	return parsed, nil
}

func GetPostIDFromUrl ( w http.ResponseWriter, r *http.Request,idPath string) (int)  {
		postID, err:=strconv.Atoi(r.PathValue("id")) 
	if err!= nil{
		writeError(w, err, http.StatusBadRequest)
		return 0
	}
	return postID
}



