package handlers

import (
	"html/template"
	"lions/internal/repository"
)

type Handler struct {
	repo      *repository.Repository
	templates map[string]*template.Template
}

func NewHandler(repo *repository.Repository, templates map[string]*template.Template) *Handler {
	return &Handler{repo: repo, templates: templates}
}
