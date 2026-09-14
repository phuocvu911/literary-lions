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
	//check seed flag
	seed := flag.Bool("seed", false, "seed the database")
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
	app := handlers.NewApp(db, templates, hasher)

	mux := http.NewServeMux()
	//register endpoints here
	mux.HandleFunc("/profile", app.ProfileHandler)
	mux.HandleFunc("/profile/upload", app.ProfileUploadHandler)
	mux.HandleFunc("/profile/update", app.ProfileUpdateHandler)
	mux.HandleFunc("/profile/image", app.ProfileImageHandler)
	
	mux.HandleFunc("/", app.NotFoundHandler) //every unregistered endpoints go here
	mux.HandleFunc("GET /register", app.HandleRegister)
	mux.HandleFunc("POST /register", app.HandleRegister)
	mux.HandleFunc("GET /login", app.HandleLogin)
	mux.HandleFunc("POST /login", app.HandleLogin)
	mux.HandleFunc("POST /logout", app.HandleLogout)
	mux.HandleFunc("GET /{$}", app.HandleHome)
	app.Router(mux)

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
