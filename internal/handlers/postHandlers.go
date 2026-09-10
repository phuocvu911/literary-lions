package handlers

import (
	"errors"
	"lions/internal/database"
	"lions/internal/models"
	"net/http"
	"strings"
	"strconv"
)

type PostListResp struct {
	Items      []models.Post `json:"items"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalItems int           `json:"totalItems"`
}

type CreatePostRequest struct {
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	CategoryIDs []int64 `json:"category_ids"`
}

type CreatePostResp struct {
	PostID int64 `json:"id"`
}

func (app *App) CreatePost(w http.ResponseWriter, r *http.Request, user *models.User) {
	var post CreatePostRequest
	if err := decodeJSON(r, &post); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	//check content is not empty
	if strings.TrimSpace(post.Content) == "" {
		writeError(w, errors.New("Comments cannot be empty"), http.StatusBadRequest)
		return
	}

	postID, err := database.CreatePost(app.db, user.ID, post.Title, post.Content, post.CategoryIDs)
	if err != nil {
		writeError(w, errors.New("failed to create post"), http.StatusInternalServerError)
		return
	}
	resp := CreatePostResp{postID}
	writeJSON(w, http.StatusCreated, resp)
}

func (app *App) ListPosts(w http.ResponseWriter, r *http.Request) {
	limit, page, offset, err := paginationParams(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	posts, err := database.ListPosts(app.db, limit, offset)
	if err != nil {
		writeError(w, errors.New("failed to load posts"), http.StatusInternalServerError)
		return
	}

	totalItems, err := database.CountPosts(app.db)
	if err != nil {
		writeError(w, errors.New("failed to count posts"), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, PostListResp{posts, page, limit, totalItems})
}

func (app *App) GetPostByID(w http.ResponseWriter, r *http.Request) {
	var post models.Post
	postIDint,err:= strconv.Atoi (r.URL.Query().Get("id"))
	if err!= nil{
		writeError(w,errors.New("id must be an int"), http.StatusBadRequest)
		return
	}
	postID := int64(postIDint)
	if post, err = database.GetPostByID(app.db, postID); err!= nil{
		writeError(w, err, http.StatusBadRequest)
	}

	writeJSON(w, http.StatusOK, post)
	
}

func (app *App) SearchPost(w http.ResponseWriter, r *http.Request) {
	keyWord := strings.TrimSpace(r.URL.Query().Get("q"))
	if keyWord == "" {
		app.ListPosts(w, r)
		return
	}

	limit, page, offset, err := paginationParams(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	posts, err := database.SearchPost(app.db, keyWord, limit, offset)
	if err != nil {
		writeError(w, errors.New("failed to search posts"), http.StatusInternalServerError)
		return
	}

	totalItems, err := database.CountSearchPosts(app.db, keyWord)
	if err != nil {
		writeError(w, errors.New("failed to count search results"), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, PostListResp{posts, page, limit, totalItems})
}
