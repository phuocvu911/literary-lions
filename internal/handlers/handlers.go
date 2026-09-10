package handlers
import (
	//"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"log"
)

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