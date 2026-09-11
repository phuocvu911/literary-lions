package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	db        *sql.DB
	templates map[string]*template.Template
	hasher    PasswordHasher
}

func NewApp(db *sql.DB, templates map[string]*template.Template, hasher PasswordHasher) *App {
	return &App{db: db, templates: templates, hasher: hasher}
}
func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("invalid JSON: " + err.Error())
	}
	return nil
}

func writeError(w http.ResponseWriter, err error, status int) {
	http.Error(w, err.Error(), status)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Println("failed to encode JSON:", err)
	}
}

func optionalPageInt(r *http.Request, name string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(name + " must be an integer")
	}
	return parsed, nil
}

// paginationParams reads page and limit query parameters and computes the SQL offset.
func paginationParams(r *http.Request) (limit, page, offset int, err error) {
	limit, err = optionalPageInt(r, "limit", 20)
	if err != nil {
		return 0, 0, 0, err
	}

	page, err = optionalPageInt(r, "page", 1)
	if err != nil {
		return 0, 0, 0, err
	}

	if limit <= 0 || limit > 100 {
		return 0, 0, 0, errors.New("limit must be between 1 and 100")
	}
	if page < 1 {
		return 0, 0, 0, errors.New("page must be at least 1")
	}

	offset = limit * (page - 1)
	return limit, page, offset, nil
}
