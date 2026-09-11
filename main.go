package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"lions/internal/database"
	"lions/internal/handlers"
	"net/http"
	"flag"
)

//go:embed internal/web
var webFS embed.FS

func main() {
	//check seed flag
	seed := flag.Bool("seed", false, "seed the database")
	flag.Parse()	

	//open db
	db, err := database.OpenDB()
	if err != nil {
		log.Fatalf("err open database: %v", err)
	}
	defer db.Close()

	//if enabled seed db
	if *seed {
		if err := database.Seed(db); err != nil {
			log.Fatal(err)
		}
	}	

	//parse templates
	templates, err := parseTemplates()
	if err != nil {
		log.Fatalf("err parse templates: %v", err)
	}

	//initialize app
	app := handlers.NewApp(db, templates)

	mux := http.NewServeMux()
	//register endpoints here
	mux.HandleFunc("/profile", app.ProfileHandler)
	mux.HandleFunc("/profile/upload", app.ProfileUploadHandler)
	mux.HandleFunc("/profile/update", app.ProfileUpdateHandler)
	mux.HandleFunc("/profile/image", app.ProfileImageHandler)
	
	//serve css
	static, err := fs.Sub(webFS, "internal/web/static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))

	//start server
	log.Printf("Literary Lions forum listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// parseTemplates builds one template set per page, each composed with the shared base layout.
func parseTemplates() (map[string]*template.Template, error) {
	//take all the html files in the pages  folder
	pages, err := fs.Glob(webFS, "internal/web/templates/pages/*.html")
	if err != nil {
		return nil, err
	}

	//parse components
	components, err := fs.Glob(webFS, "internal/web/templates/components/*.html")
	if err != nil {
		return nil, err
	}

	files := append(
		[]string{"internal/web/templates/base.html"},
		components...,
	)

	base, err := template.ParseFS(webFS, files...)
	if err != nil {
		return nil, err
	}

	templates := make(map[string]*template.Template, len(pages))

	// for each page, parse the base layout and the page template together
	for _, page := range pages {
		name := page[len("internal/web/templates/pages/"):]
		t, err := base.Clone()
		if err != nil {
			return nil, err
		}

		t, err = t.ParseFS(webFS, page)
		if err != nil {
			return nil, err
		}

		templates[name] = t
	}
	return templates, nil
}

// render executes a page template, moved to helper.go in package handlers,
