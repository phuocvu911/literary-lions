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

// NewPostForm renders the page where a signed-in user writes a post.
func (app *App) NewPostForm(w http.ResponseWriter, r *http.Request, user *models.User) {
	categories, err := database.ListCategories(app.db)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, "newpost.html", map[string]any{
		"User":       user,
		"Categories": categories,
	})
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

// CreatePostFromForm creates a post submitted by a normal HTML form.
// It is separate from CreatePost because CreatePost accepts a JSON request body.
func (app *App) CreatePostFromForm(w http.ResponseWriter, r *http.Request, user *models.User) {
	if err := r.ParseForm(); err != nil {
		writeError(w, errors.New("invalid form data"), http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categoryValues := r.Form["category_ids"]
	categoryIDs := make([]int64, 0, len(categoryValues))

	for _, value := range categoryValues {
		categoryID, err := strconv.ParseInt(value, 10, 64)
		if err != nil || categoryID <= 0 {
			writeError(w, errors.New("invalid category"), http.StatusBadRequest)
			return
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	if title == "" {
		writeError(w, errors.New("title cannot be empty"), http.StatusBadRequest)
		return
	}
	if content == "" {
		writeError(w, errors.New("content cannot be empty"), http.StatusBadRequest)
		return
	}
	if len(categoryIDs) == 0 {
		writeError(w, errors.New("at least one category is required"), http.StatusBadRequest)
		return
	}

	if _, err := database.CreatePost(app.db, user.ID, title, content, categoryIDs); err != nil {
		writeError(w, errors.New("failed to create post"), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
