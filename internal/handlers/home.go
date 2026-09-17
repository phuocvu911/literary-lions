package handlers

import (
	"lions/internal/database"
	"net/http"
)

func (app *App) HandleHome(w http.ResponseWriter, r *http.Request) {
	user := app.currentUser(r)
	posts, err := database.ListPosts(app.db, 20, 0)
	if err != nil {
		app.serverError(w, err)
		return
	}

	categories, err := database.ListCategories(app.db)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, "home.html", map[string]any{
		"User":       user,
		"Posts":      posts,
		"Categories": categories,
	})
}
