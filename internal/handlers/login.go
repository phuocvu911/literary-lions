package handlers

import (
	"database/sql"
	"lions/internal/database"
	"net/http"
	"strings"
)

// login return same msg in every unsucessful login attempt
func (app *App) HandleLogin(w http.ResponseWriter, r *http.Request) {
	//check if user already logged in
	if app.currentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//if GET, serve the template
	if r.Method == http.MethodGet {
		app.render(w, "login.html", nil)
		return
	}

	//get input
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	//fail scenario closure
	fail := func() {
		w.WriteHeader(http.StatusUnauthorized)
		app.render(w, "login.html", map[string]string{
			"Error": "Email or password is incorrect.",
		})
	}

	//check email and password match
	user, err := database.UserByEmail(app.db, email)
	if err == nil {
		if isRightPassword := app.hasher.Verify(password, user.PasswordHash); !isRightPassword {
			fail()
			return
		}
	}
	if err != nil {
		if err != sql.ErrNoRows {
			app.serverError(w, err)
			return
		}
		fail()
		return
	}

	//start sessions
	if err := app.startSession(w, user.ID); err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
