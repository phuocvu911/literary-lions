package handlers

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher swaps implementations without touching calling code
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) bool
}

// SHA-256
type SHA256Hasher struct{}

func (SHA256Hasher) Hash(password string) (string, error) {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:]), nil
}

func (SHA256Hasher) Verify(password, hash string) bool {
	computed, _ := SHA256Hasher{}.Hash(password)
	return computed == hash
}

// bcrypt

type BcryptHasher struct {
	Cost int
}

func NewBcryptHasher() BcryptHasher {
	return BcryptHasher{Cost: bcrypt.DefaultCost}
}

func (b BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.Cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (BcryptHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
