package handlers

import (
	"lions/internal/models"
	"net/http"
)

type NotFoundData struct {
	Page string
	User *models.User
}

func (a *App) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	a.render(w, "404.html", NotFoundData{Page: "404"})
}
