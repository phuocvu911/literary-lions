package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"lions/internal/models"
	"log"
	"movies-api/models"
	"movies-api/service"
	"net/http"
	"strconv"
)


func (app *App) CreateComment(w http.ResponseWriter, r *http.Request, user *models.User) {
	var comment models.Comment
	if err := decodeJSON(r, &comment); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	created, err := app.db.CreateComment()
}

func (app *App)ListComment(w http.ResponseWriter, r *http.Request) {

}

