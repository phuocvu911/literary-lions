package main

import (
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"lions/internal/database"
	"lions/internal/handlers"
	"log"
	"net/http"
)

//go:embed internal/web
var webFS embed.FS

func main() {
	//parse flag
	useBcrypt := flag.Bool("bcrypt", false, "use bcrypt instead of sha256 for password hashing")
	flag.Parse()

	var hasher handlers.PasswordHasher
	if *useBcrypt {
		hasher = handlers.NewBcryptHasher()
		log.Println("using bcrypt for password hashing")
	} else {
		hasher = handlers.SHA256Hasher{}
		log.Println("using sha256 for password hashing")
	}
	//open db
	db, err := database.OpenDB()
	if err != nil {
		log.Fatalf("err open database: %v", err)
	}
	defer db.Close()

	//parse templates
	templates, err := parseTemplates()
	if err != nil {
		log.Fatalf("err parse templates: %v", err)
	}

	//initialize app
	app := handlers.NewApp(db, templates, hasher)

	mux := http.NewServeMux()
	//register endpoints here
	mux.HandleFunc("/", app.NotFoundHandler) //every unregistered endpoints go here
	mux.HandleFunc("GET /register", app.HandleRegister)
	mux.HandleFunc("POST /register", app.HandleRegister)
	mux.HandleFunc("GET /login", app.HandleLogin)
	mux.HandleFunc("POST /login", app.HandleLogin)

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

// render executes a page template, moved to helper.go in package handlers,
