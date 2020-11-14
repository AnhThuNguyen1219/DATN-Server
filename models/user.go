package models

import (
	"html"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
}
type UserInfo struct {
	ID        int
	Username  string
	AvatarURL string
	DOB       string
}

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hashedPassword, password string) bool {
	if hashedPassword != password {
		return false
	}
	return true
}

func Santize(data string) string {
	data = html.EscapeString(strings.TrimSpace(data))
	return data
}
