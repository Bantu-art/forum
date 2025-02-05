package utils

import (
	"regexp"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordsHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

func ValidateUsername(username string) bool {
	return len(username) >= 3 && len(username) <= 30
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasLower := false
	hasUpper := false
	hasNumber := false
	hasSpecial := false
	for _, char := range password {
		switch {
		case 'a' <= char && char <= 'z':
			hasLower = true
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '@' || char == '$' || char == '!' || char == '%' || char == '*' || char == '?' || char == '&':
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasNumber && hasSpecial
}

func GenerateId() string {
	Uid, _ := uuid.NewV4()
	return Uid.String()
}
