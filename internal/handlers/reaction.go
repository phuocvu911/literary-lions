package handlers

import (
	"net/http"
	"encoding/json"
	"lions/internal/models"
	"strconv"
	"database/sql"
	"fmt"
)

type ReactionRequest struct {
	Value int `json:"value"`
}

type ReactionResponse struct {
	Likes        int `json:"likes"`
	Dislikes     int `json:"dislikes"`
	ReactionToPost int `json:"reaction_to_post"`
	ReactionToComment int `json:"reaction_to_comment"`
}

func (app *App) ReactToPost(w http.ResponseWriter, r *http.Request, user *models.User) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	var req ReactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Value != 1 && req.Value != -1 {
		http.Error(w, "invalid reaction", http.StatusBadRequest)
		return
	}

	// Check the user's current reaction
	var currentReaction int

	err = app.db.QueryRow(`
		SELECT value
		FROM post_reactions
		WHERE user_id = ? AND post_id = ?
	`, user.ID, postID).Scan(&currentReaction)

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "failed to get current reaction", http.StatusInternalServerError)
		return
	}

	if currentReaction == req.Value {
		// Same button clicked again -> remove reaction
		_, err = app.db.Exec(`
			DELETE FROM post_reactions
			WHERE user_id = ? AND post_id = ?
		`, user.ID, postID)
	} else {
		// No reaction or opposite reaction -> set/switch reaction
		_, err = app.db.Exec(`
			INSERT INTO post_reactions (
				user_id,
				post_id,
				value
			)
			VALUES (?, ?, ?)
			ON CONFLICT(user_id, post_id)
			DO UPDATE SET value = excluded.value
		`, user.ID, postID, req.Value)
	}

	if err != nil {
		http.Error(w, "failed to save reaction", http.StatusInternalServerError)
		return
	}

	// Get updated counts
	var likes int
	var dislikes int

	err = app.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END), 0)
		FROM post_reactions
		WHERE post_id = ?
	`, postID).Scan(&likes, &dislikes)

	if err != nil {
		http.Error(w, "failed to get reaction counts", http.StatusInternalServerError)
		return
	}

	// Get the user's new reaction
	var reactionToPost int

	err = app.db.QueryRow(`
		SELECT value
		FROM post_reactions
		WHERE user_id = ? AND post_id = ?
	`, user.ID, postID).Scan(&reactionToPost)

	if err == sql.ErrNoRows {
		reactionToPost = 0
	} else if err != nil {
		http.Error(w, "failed to get user reaction", http.StatusInternalServerError)
		return
	}

	response := ReactionResponse{
		Likes:        likes,
		Dislikes:     dislikes,
		ReactionToPost: reactionToPost,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) ReactToComment(w http.ResponseWriter, r *http.Request, user *models.User) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("handler fired")

	commentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	var req ReactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Value != 1 && req.Value != -1 {
		http.Error(w, "invalid reaction", http.StatusBadRequest)
		return
	}

	// Check the user's current reaction
	var currentReaction int

	err = app.db.QueryRow(`
		SELECT value
		FROM comment_reactions
		WHERE user_id = ? AND comment_id = ?
	`, user.ID, commentID).Scan(&currentReaction)

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "failed to get current reaction", http.StatusInternalServerError)
		return
	}

	if currentReaction == req.Value {
		// Same button clicked again -> remove reaction
		_, err = app.db.Exec(`
			DELETE FROM comment_reactions
			WHERE user_id = ? AND comment_id = ?
		`, user.ID, commentID)
	} else {
		// No reaction or opposite reaction -> set/switch reaction
		_, err = app.db.Exec(`
			INSERT INTO comment_reactions (
				user_id,
				comment_id,
				value
			)
			VALUES (?, ?, ?)
			ON CONFLICT(user_id, comment_id)
			DO UPDATE SET value = excluded.value
		`, user.ID, commentID, req.Value)
	}

	if err != nil {
		http.Error(w, "failed to save reaction", http.StatusInternalServerError)
		return
	}

	// Get updated counts
	var likes int
	var dislikes int

	err = app.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN value = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN value = -1 THEN 1 ELSE 0 END), 0)
		FROM comment_reactions
		WHERE comment_id = ?
	`, commentID).Scan(&likes, &dislikes)

	if err != nil {
		http.Error(w, "failed to get reaction counts", http.StatusInternalServerError)
		return
	}

	// Get the user's new reaction
	var reactionToComment int

	err = app.db.QueryRow(`
		SELECT value
		FROM comment_reactions
		WHERE user_id = ? AND comment_id = ?
	`, user.ID, commentID).Scan(&reactionToComment)

	if err == sql.ErrNoRows {
		reactionToComment = 0
	} else if err != nil {
		http.Error(w, "failed to get user reaction", http.StatusInternalServerError)
		return
	}

	response := ReactionResponse{
		Likes:        likes,
		Dislikes:     dislikes,
		ReactionToComment: reactionToComment,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}