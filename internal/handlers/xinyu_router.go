package handlers

import "net/http"

func (app *App) Router(mux *http.ServeMux) {
	// Public forum routes
	mux.Handle("GET /posts", app.ListPosts)
	mux.Handle("GET /posts/search", app.SearchPost)
	mux.Handle("GET /posts/{id}", app.GetPostByID)
	mux.Handle("GET /categories", app.ListCategories)
	mux.Handle("GET /posts/{id}/comments", app.ListComment)

	// Routes that require a valid login session.
	mux.Handle("POST /posts", app.RequireAuth(app.CreatePost))
	mux.Handle("POST /posts/{id}/comments", app.RequireAuth(app.CreateComment))
}
