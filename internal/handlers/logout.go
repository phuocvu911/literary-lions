package handlers

import (
	"lions/internal/database"
	"net/http"
)

func (app *App) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// retrieve the session from client side
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		// if so, delete it
		if err := database.DeleteSession(app.db, cookie.Value); err != nil {
			app.serverError(w, err)
			return
		}
	}
	//delete session from client side
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
