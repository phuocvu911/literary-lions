package handlers

import "net/http"

type NotFoundData struct {
	Page string
}

func (a *App)NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	a.render(w, "404.html", NotFoundData{Page: "404"})
}
