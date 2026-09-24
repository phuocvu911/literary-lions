package handlers

import (
	"lions/internal/database"
	"lions/internal/models"
	"net/http"
	"time"
	"strings"
)

type FullPost struct {
	ID            int64
	Title         string
	Content       string
	CreatedAt     time.Time
	Author        string
	Likes         int
	Dislikes      int
	CommentCount  int
	CategoryNames string
	Categories    []models.Category
	CurrentReaction int
}

type PostWithReaction struct {
	UserID int
	PostID int64
	Value int
}

func (app *App) HandleHome(w http.ResponseWriter, r *http.Request) {
	user := app.currentUser(r)
	loggedIn := false
	if user != nil {
		loggedIn = true
	}

	query := r.URL.Query()

	// Get filters from URL
	filters := database.PostFilters{
		Keyword:  strings.TrimSpace(query.Get("q")),
		Category: query.Get("category"),
		Sort:     query.Get("sort"),
		Time:     query.Get("time"),
	}	

	posts, err := database.ListPosts(app.db, filters, 20, 0)
	if err != nil {
		app.serverError(w, err)
		return
	}

	//all post reactions
	reactionsToPosts := make(map[int64]int)

	if loggedIn {
		rows, err:= app.db.Query(`
			SELECT user_id, post_id, value
			FROM post_reactions
			WHERE user_id = ?
		`, app.currentUser(r).ID)

		if err != nil {
			return
		}
		defer rows.Close()

		var postsWithReaction []PostWithReaction
		for rows.Next() {
			p := &PostWithReaction{}
			err := rows.Scan(&p.UserID, &p.PostID, &p.Value)
			if err != nil {
				return
			}
			postsWithReaction = append(postsWithReaction, *p)
		}		

		for _, p := range postsWithReaction {
			reactionsToPosts[p.PostID] = p.Value
		}
	}

	var fullposts []FullPost
	for _, p := range posts {
		fullpost := FullPost{
			ID: p.ID, 
			Title: p.Title, 
			Content: p.Content,
			CreatedAt: p.CreatedAt, 
			Author: p.Author,
			Likes: p.Likes,
			Dislikes: p.Dislikes,
			CommentCount: p.CommentCount,
			CategoryNames: p.CategoryNames,
			Categories: p.Categories,
			CurrentReaction: reactionsToPosts[p.ID],
		}
		fullposts = append(fullposts, fullpost)	
	}

	categories, err := database.ListCategories(app.db)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, "home.html", map[string]any{
		"User":       user,
		"LoggedIn": loggedIn,		
		"Posts":      fullposts,
		"Categories": categories,
	})
}
