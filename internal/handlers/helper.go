package handlers

import (
	"lions/internal/database"
	"lions/internal/models"
	"log"
	"net/http"
	"time"
)

const (
	sessionCookie = "lions_session"
	sessionMaxAge = 24 * time.Hour
	maxCredentials = 100 // max length for email/username/password fields
)

// render execute a page template
func (a *App) render(w http.ResponseWriter, page string, data any) {
	t, ok := a.templates[page]
	if !ok {
		log.Printf("Template %s not found in cache", page)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("Error rendering template %s: %v", page, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// currentUser resolves the logged-in user from the session cookie, or nil.
func (app *App) currentUser(r *http.Request) *models.User {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil
	}
	user, err := database.UserBySession(app.db, cookie.Value)
	if err != nil {
		return nil
	}
	return user
}

// use when need to return code 500
func (app *App) serverError(w http.ResponseWriter, err error) {
	log.Printf("server error: %v", err)
	http.Error(w, "Sorry, something went wrong.", http.StatusInternalServerError)
}
