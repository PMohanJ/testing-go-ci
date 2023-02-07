package helpers

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashPassowrd(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Panic("err generating a hashed password: ", err)
	}
	return string(bytes)
}

func VerifyPassword(hashedPassword string, givenPassowrd string) (string, bool) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(givenPassowrd))
	isValid := true

	if err != nil {
		isValid = false
		return fmt.Sprintf("Given password is invalid"), isValid
	}
	return "", isValid
}
