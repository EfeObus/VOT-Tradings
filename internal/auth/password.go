// Package auth implements password hashing and session management for
// VOT Tradings user accounts.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword returns a bcrypt hash suitable for storing in
// users.password_hash.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword reports whether password matches the stored bcrypt hash.
func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
