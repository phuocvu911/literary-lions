package handlers

import (
	"database/sql"
	"html/template"
)

type App struct {
	db        *sql.DB
	templates map[string]*template.Template
	hasher    PasswordHasher
}

func NewApp(db *sql.DB, templates map[string]*template.Template, hasher PasswordHasher) *App {
	return &App{db: db, templates: templates, hasher: hasher}
}
