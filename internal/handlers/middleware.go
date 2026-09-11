package handlers

import (
	"lions/internal/models"
	"net/http"
)

// RequireAuth wraps a handler so guests are redirected to the login page.
func (app *App) RequireAuth(next func(http.ResponseWriter, *http.Request, *models.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := app.currentUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r, user)
	}
}
