package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/smtp"
	"time"
)

const tokenExpirationTime = 30 * time.Minute

// generateToken returns a random 32-byte hex string
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HandleForgotPassword handles GET (show form) and POST (issue token + email it).
func (a *App) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	//if GET, serve the page
	if r.Method == http.MethodGet {
		a.render(w, "forgot_password.html", nil)
		return
	}

	email := r.FormValue("email")

	var userID int
	err := a.db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)

	// don't reveal whether the email exists.
	if err != nil {
		a.render(w, "forgot_password_sent.html", nil)
		return
	}

	token, err := generateToken()
	if err != nil {
		a.serverError(w, err)
		return
	}

	expiresAt := time.Now().Add(tokenExpirationTime)
	_, err = a.db.Exec(
		"INSERT INTO password_resets (token, user_id, expires_at, used) VALUES (?, ?, ?, 0)",
		token, userID, expiresAt,
	)
	if err != nil {
		a.serverError(w, err)
		return
	}

	if err := sendResetEmail(email, token); err != nil {
		// Log it, but still show the generic success page — don't expose email delivery failures to the client.
		fmt.Printf("failed to send reset email: %v\n", err)
	}

	a.render(w, "forgot_password_sent.html", nil)
}

// sendResetEmail sends the reset link via SMTP.
func sendResetEmail(toEmail, token string) error {
	from := "iamphuocvux@gmail.com" //placeholder email
	password := "woxz ldnk gxaz fhlu"

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	resetLink := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	msg := []byte(
		"From: Literary Lions Forum\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: Reset your Literary Lions password\r\n" +
			"Mime-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
			"Click the link below to reset your password. This link expires in 30 minutes.\r\n" +
			resetLink + "\r\n",
	)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, msg)
}

// handleResetPassword handles GET (show new-password form) and POST (apply it).
func (a *App) HandleResetPassword(w http.ResponseWriter, r *http.Request) {

	token := r.URL.Query().Get("token")
	if r.Method == http.MethodPost {
		token = r.FormValue("token")
	}

	var userID int
	var expiresAt time.Time
	var used bool
	err := a.db.QueryRow(
		"SELECT user_id, expires_at, used FROM password_resets WHERE token = ?",
		token,
	).Scan(&userID, &expiresAt, &used)

	if err != nil || used || time.Now().After(expiresAt) {
		http.Error(w, "This reset link is invalid or has expired.", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		a.render(w, "reset_password.html", map[string]string{"Token": token})
		return
	}

	newPassword := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if len(newPassword) < 8 {
		a.render(w, "reset_password.html", map[string]string{
			"Token": token,
			"Error": "Password must be at least 8 characters long.",
		})
		return
	}

	if newPassword != confirmPassword {
		a.render(w, "reset_password.html", map[string]string{
			"Token": token,
			"Error": "Passwords do not match.",
		})
		return
	}
	hashed, err := a.hasher.Hash(newPassword)
	if err != nil {
		a.serverError(w, err)
		return
	}

	tx, err := a.db.Begin()

	if err != nil {
		a.serverError(w, err)
		return
	}
	defer tx.Rollback()

	//update new password_hash
	if _, err := tx.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hashed, userID); err != nil {
		a.serverError(w, err)
		return
	}

	//mark the token as used
	if _, err := tx.Exec("UPDATE password_resets SET used = 1 WHERE token = ?", token); err != nil {
		a.serverError(w, err)
		return
	}

	//kill any active sessions for this user so a stolen session doesn't survive a password reset.
	_, _ = tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID)

	//dont delete the row in password_reset for sucessful password change so we can audit

	if err := tx.Commit(); err != nil {
		a.serverError(w, err)
		return
	}

	a.render(w, "reset_password.html", map[string]string{
		"Success": "Password reset successfully!",
	})
}
