package handlers

import (
	"database/sql"
	"errors"
	"lions/internal/database"
	"lions/internal/models"
	"net/http"
	"strconv"
	"strings"
)

type CommentListResponse struct {
	Items      []models.Comment `json:"items"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalItems int              `json:"totalItems"`
}

func (app *App) CreateComment(w http.ResponseWriter, r *http.Request, user *models.User) {
	var comment models.Comment
	if err := decodeJSON(r, &comment); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	//check content is not empty
	if strings.TrimSpace(comment.Content) == "" {
		writeError(w, errors.New("Comments cannot be empty"), http.StatusBadRequest)
		return
	}
	postID, err := GetPostIDFromURL(r, "id")
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	if _, err := database.GetPostByID(app.db, postID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, errors.New("post not found"), http.StatusNotFound)
			return
		}
		writeError(w, errors.New("failed to load post"), http.StatusInternalServerError)
		return
	}

	commentID, err := database.CreateComment(app.db, user.ID, postID, comment.Content)
	if err != nil {
		writeError(w, errors.New("failed to create comment"), http.StatusInternalServerError)
		return
	}
	comment.ID = commentID

	writeJSON(w, http.StatusCreated, comment)
}

func (app *App) ListComment(w http.ResponseWriter, r *http.Request) {
	limit, page, offset, err := paginationParams(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	postID, err := GetPostIDFromURL(r, "id")
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	if _, err := database.GetPostByID(app.db, postID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, errors.New("post not found"), http.StatusNotFound)
			return
		}
		writeError(w, errors.New("failed to load post"), http.StatusInternalServerError)
		return
	}

	comments, err := database.ListComments(app.db, postID, limit, offset)
	if err != nil {
		writeError(w, errors.New("failed to load comments"), http.StatusInternalServerError)
		return
	}

	totalItems, err := database.CountComments(app.db, postID)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, CommentListResponse{comments, page, limit, totalItems})

}

func GetPostIDFromURL(r *http.Request, idPath string) (int64, error) {
	postID, err := strconv.ParseInt(r.PathValue(idPath), 10, 64)
	if err != nil {
		return 0, errors.New("post ID must be an integer")
	}
	if postID <= 0 {
		return 0, errors.New("post ID must be positive")
	}

	return postID, nil
}
