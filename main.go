package main

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"lions/internal/database"
	"log"
	"net/http"
)

//go:embed internal/web
var webFS embed.FS

// so our backend just a db and web, and handler will be the method of app
type App struct {
	db        *sql.DB
	templates map[string]*template.Template
}

func main() {
	db, err := database.OpenDB()
	if err != nil {
		log.Fatalf("err open database: %v", err)
	}
	defer db.Close()

	templates, err := parseTemplates()
	if err != nil {
		log.Fatalf("err parse templates: %v", err)
	}
	app := &App{
		db:        db,
		templates: templates,
	}

	mux := http.NewServeMux()
	//register endpoints here

	static, err := fs.Sub(webFS, "web/static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))

	log.Printf("Literary Lions forum listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// parseTemplates builds one template set per page, each composed with the
// shared base layout.
func parseTemplates() (map[string]*template.Template, error) {
	//take all the html files in the pages  folder
	pages, err := fs.Glob(webFS, "internal/web/templates/pages/*.html")
	if err != nil {
		return nil, err
	}
	templates := make(map[string]*template.Template, len(pages))

	// for each page, parse the base layout and the page template together
	for _, page := range pages {
		name := page[len("internal/web/templates/pages/"):]
		t, err := template.ParseFS(webFS, "internal/web/templates/base.html", page)
		if err != nil {
			return nil, err
		}
		templates[name] = t
	}
	return templates, nil
}

// render executes a page template with shared data (current user) merged in.
func (app *App) render(w http.ResponseWriter, page string, data any) {
	t, ok := app.templates[page]
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
