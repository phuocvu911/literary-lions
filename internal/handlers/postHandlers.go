package handlers

import (
	"database/sql"
	"errors"
	"lions/internal/database"
	"lions/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"
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

type FullComment struct {
	ID        int64
	Content   string
	CreatedAt time.Time
	Author    string
	Likes     int
	Dislikes  int	
	CurrentReaction int
	CommentFiles []CommentFile
}

type CommentFile struct {
	ID          int64
	Filename    string
	ContentType string
}

type CommentWithReaction struct {
	UserID int
	CommentID int64
	Value int
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
    query := r.URL.Query()

	//get data from filters
    filters := database.PostFilters{
        Keyword:  strings.TrimSpace(query.Get("q")),
        Category: query.Get("category"),
        Sort:     query.Get("sort"),
        Time:     query.Get("time"),
    }

	limit, page, offset, err := paginationParams(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	posts, err := database.ListPosts(app.db, filters, limit, offset)
	if err != nil {
		writeError(w, errors.New("failed to load posts"), http.StatusInternalServerError)
		return
	}

	totalItems, err := database.CountPosts(app.db, filters)
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

// PostPage renders one post and its comments for browser visitors.
func (app *App) PostPage(w http.ResponseWriter, r *http.Request) {
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
		app.serverError(w, err)
		return
	}

	comments, err := database.ListComments(app.db, postID, 100, 0)
	if err != nil {
		app.serverError(w, err)
		return
	}

	//current reaction to the post
	reactionToPost, err := app.getReactionToPost(r, post.ID)
	if err != nil {
		app.serverError(w, err)
		return		
	}		

	fullComments := []FullComment{}

	if app.currentUser(r) != nil {
		//map commentFiles
		filesToComments, err := app.getFilesToComments(comments)
		if err != nil {
			app.serverError(w, err)
			return		
		}

		//current reactions to the comments
		reactionsToComments, err := app.getReactionsToComments(r)
		if err != nil {
			app.serverError(w, err)
			return		
		}		

		for _, c := range comments {
			fullComment := FullComment{
				ID: c.ID,
				Content: c.Content,
				CreatedAt: c.CreatedAt, 
				Author: c.Author,
				Likes: c.Likes,
				Dislikes: c.Dislikes,
				CurrentReaction: reactionsToComments[c.ID],
				CommentFiles: filesToComments[c.ID],
			}
			fullComments = append(fullComments, fullComment)
		}
	} else {
		for _, c := range comments {
			fullComment := FullComment{
				ID: c.ID,
				Content: c.Content,
				CreatedAt: c.CreatedAt, 
				Author: c.Author,
				Likes: c.Likes,
				Dislikes: c.Dislikes,
			}
			fullComments = append(fullComments, fullComment)
		}		
	}

	app.render(w, "post.html", map[string]any{
		"User":     app.currentUser(r),
		"Post":     post,
		"Comments": fullComments,
		"ReactionToPost": reactionToPost,
	})
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

//helpers 
func (app *App) getFilesToComments(comments []models.Comment) (map[int64][]CommentFile, error) {
	filesToComments := make(map[int64][]CommentFile)

	for _, comment := range comments {
		rows, err := app.db.Query(`
			SELECT id, filename, content_type
			FROM files
			WHERE comment_id = ?
		`, comment.ID)

		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var files []CommentFile

		for rows.Next() {
			var file CommentFile

			if err := rows.Scan(&file.ID, &file.Filename, &file.ContentType); err != nil {
				return nil, err
			}
			files = append(files, file)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}		

		filesToComments[comment.ID] = files
	}

	return filesToComments, nil
}

func (app *App) getReactionsToComments(r *http.Request) (map[int64]int, error) {
	reactionsToComments := make(map[int64]int)

	rows, err:= app.db.Query(`
		SELECT user_id, comment_id, value
		FROM comment_reactions
		WHERE user_id = ?
	`, app.currentUser(r).ID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commentsWithReaction []CommentWithReaction
	for rows.Next() {
		c := &CommentWithReaction{}
		err := rows.Scan(&c.UserID, &c.CommentID, &c.Value)
		if err != nil {
			return nil, err
		}
		commentsWithReaction = append(commentsWithReaction, *c)
	}		

	for _, c := range commentsWithReaction {
		reactionsToComments[c.CommentID] = c.Value
	}

	return reactionsToComments, nil
}

func (app *App) getReactionToPost(r *http.Request, postID int64) (int, error){
	var reactionToPost int

	if app.currentUser(r) != nil {
		err := app.db.QueryRow(`
			SELECT value
			FROM post_reactions
			WHERE user_id = ? AND post_id = ?
		`, app.currentUser(r).ID, postID).Scan(&reactionToPost)

		if err == sql.ErrNoRows {
			reactionToPost = 0
		} else if err != nil {
			return 0, err
		}	
	}	

	return reactionToPost, nil
}