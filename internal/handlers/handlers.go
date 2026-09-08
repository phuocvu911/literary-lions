package handlers

import (
	"database/sql"
	"html/template"
)

type App struct {
	db        *sql.DB
	templates map[string]*template.Template
}

func NewApp(db *sql.DB, templates map[string]*template.Template) *App {
	return &App{db: db, templates: templates}
}
