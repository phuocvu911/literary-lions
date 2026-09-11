package handlers

import (
	"net/http"
)

func (app *App) HandleHome(w http.ResponseWriter, r *http.Request) {
	user := app.currentUser(r)

	//do something

	app.render(w, "home.html", map[string]any{
		"User": user,
	})
}
