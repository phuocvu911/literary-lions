package handlers

import "net/http"


func (app *App) Router(mux *http.ServeMux) {
	mux.Handle("POST /posts/{id}/comments",app.RequireAuth(app.CreateComment))
	mux.Handle("GET /posts/{id}/comments", app.ListComment)
}
