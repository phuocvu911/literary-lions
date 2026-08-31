package main

import (
	"database/sql"
	"html/template"
	"lions/internal/database"
	"log"
)

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
}
