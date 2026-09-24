package handlers

import (
	"net/http"
	"database/sql"
	"strconv"	
	"errors"
)

func (app *App) File(w http.ResponseWriter, r *http.Request) {
	fileID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid file ID", http.StatusBadRequest)
		return
	}

	var (
		data        []byte
		filename    string
		contentType string
	)

	err = app.db.QueryRow(`
		SELECT data, filename, content_type
		FROM files
		WHERE id = ?
	`, fileID).Scan(&data, &filename, &contentType)

	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		app.serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", `inline; filename="`+filename+`"`)
	w.Write(data)
}