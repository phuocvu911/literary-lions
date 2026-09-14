package handlers

import (
	"database/sql"
	"errors"
	"lions/internal/database"
	"lions/internal/models"
	"net/http"
	"strings"
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
	post.Title = strings.TrimSpace(post.Title)
	post.Content = strings.TrimSpace(post.Content)

	if post.Title == "" {
		writeError(w, errors.New("title cannot be empty"), http.StatusBadRequest)
		return
	}
	if post.Content == "" {
		writeError(w, errors.New("content cannot be empty"), http.StatusBadRequest)
		return
	}
	if len(post.CategoryIDs) == 0 {
		writeError(w, errors.New("at least one category is required"), http.StatusBadRequest)
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
	postID, err := GetPostIDFromURL(r, "id")
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	post, err := database.GetPostByID(app.db, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, errors.New("post not found"), http.StatusNotFound)
			return
		}
		writeError(w, errors.New("failed to load post"), http.StatusInternalServerError)
		return
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
