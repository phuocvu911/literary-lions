package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
)

type App struct {
	db        *sql.DB
	templates map[string]*template.Template
}

type PageData struct {
    User *User
}

type User struct {
    Username string
}

func NewApp(db *sql.DB, templates map[string]*template.Template) *App {
	return &App{db: db, templates: templates}
}

func (app *App) ProfileHandler(w http.ResponseWriter, r *http.Request) {
    data := PageData{
        User: &User{
            Username: "John Doe",
        },
    }

    app.render(w, "profile.html", data)
}