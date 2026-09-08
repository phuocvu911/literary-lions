package handlers

import (
	"log"
	"net/http"
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
