package handlers

import (
	"crypto/subtle"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Hash password for new user accounts
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// Verify password during login
func CheckPassword(password string, storedPassword string) error {

	// Check whether stored password is bcrypt
	if strings.HasPrefix(storedPassword, "$2a$") ||
		strings.HasPrefix(storedPassword, "$2b$") ||
		strings.HasPrefix(storedPassword, "$2y$") {

		return bcrypt.CompareHashAndPassword(
			[]byte(storedPassword),
			[]byte(password),
		)
	}

	// Temporary support for existing plain-text accounts
	if subtle.ConstantTimeCompare(
		[]byte(password),
		[]byte(storedPassword),
	) == 1 {
		return nil
	}

	return bcrypt.ErrMismatchedHashAndPassword
}
