package password

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the given password using bcrypt
func HashPassword(password string) (string, error) {
	// Use cost 12 for better security balance
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

// CheckPassword checks if the provided password matches the hashed password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
