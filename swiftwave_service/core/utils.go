package core

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// SetPassword : set password for user
func (user *User) SetPassword(password string) error {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}
	user.PasswordHash = string(hashedPasswordBytes)
	return nil
}

// CheckPassword : check password for user
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// GenerateJWT : generate jwt token for user
func (user *User) GenerateJWT(jwtSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"nbf":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
		"iat":      time.Now().Unix(),
		"iss":      "swiftwave",
		"username": user.Username,
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(jwtSecret)
}
