package handlers

import "net/http"

func (app *App) Router(mux *http.ServeMux) {
	// Public forum routes
	mux.HandleFunc("GET /posts", app.ListPosts)
	mux.HandleFunc("GET /posts/search", app.SearchPost)
	mux.HandleFunc("GET /posts/{id}", app.GetPostByID)
	mux.HandleFunc("GET /categories", app.ListCategories)
	mux.HandleFunc("GET /posts/{id}/comments", app.ListComment)

	// Routes that require a valid login session.
	mux.Handle("POST /posts", app.RequireAuth(app.CreatePost))
	mux.Handle("POST /posts/{id}/comments", app.RequireAuth(app.CreateComment))
}
