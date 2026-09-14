package handlers

import (
	"lions/internal/database"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"uuid"
)

func (app *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
	//check if user already logged in
	if app.currentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//if GET, serve the page
	if r.Method == http.MethodGet {
		app.render(w, "register.html", nil)
		return
	}

	//take input
	email := strings.TrimSpace(r.FormValue("email"))
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	//fail scenarios closure
	fail := func(msg string) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		app.render(w, "register.html", map[string]string{
			"Error": msg, "Email": email, "Username": username,
		})
	}

	//validate input
	if _, err := mail.ParseAddress(email); err != nil || len(email) > maxCredentials {
		fail("Please enter a valid email address.")
		return
	}
	if len(username) < 3 || len(username) > maxCredentials || strings.ContainsAny(username, " \t\v\n@") {
		fail("Username must be at least 3 characters, with no spaces or '@'.")
		return
	}
	if len(password) < 8 || len(password) > maxCredentials {
		fail("Password must be at least 8 characters.")
		return
	}

	//hash password
	hash, err := app.hasher.Hash(password)
	if err != nil {
		app.serverError(w, err)
		return
	}

	//write to db, also see if write is possible
	userID, err := database.CreateUser(app.db, email, username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			fail("Email address is already registered.")
		} else if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			fail("Username is already taken.")
		} else {
			app.serverError(w, err)
		}
		return
	}

	//start session and redirect user to homepage, in successful path
	if err := app.startSession(w, userID); err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// startSession issues a UUID session, stores it, and sets the cookie.
func (app *App) startSession(w http.ResponseWriter, userID int64) error {
	sessionID := uuid.NewV4().String()
	expires := time.Now().Add(sessionMaxAge)
	if err := database.CreateSession(app.db, sessionID, userID, expires); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    sessionID,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}
