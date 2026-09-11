package handlers

import (
	//"database/sql"
	"errors"
	"lions/internal/database"
	//"lions/internal/models"
	"net/http"

)


func (app *App) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := database.ListCategories(app.db)
	if err != nil {
		writeError(w, errors.New("failed to load categories"), http.StatusInternalServerError)
		return
	}



	writeJSON(w, http.StatusOK, categories)


}