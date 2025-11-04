// pkg/utils/hash.go
package utils

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain text password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash verifies a password against its hash
func CheckPasswordHash(password, hash string) bool {
	// Debug logging
	log.Printf("DEBUG CheckPasswordHash:")
	log.Printf("  - Password length: %d", len(password))
	log.Printf("  - Hash length: %d", len(hash))
	log.Printf("  - Hash: %s", hash)

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	if err != nil {
		log.Printf("  - bcrypt.CompareHashAndPassword ERROR: %v", err)
	} else {
		log.Printf("  - bcrypt.CompareHashAndPassword SUCCESS")
	}

	return err == nil
}
