// Package password :: https://golangbyexample.com/generate-random-password-golang/
package main

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strings"
)

const (
	bcryptCost = bcrypt.MinCost
)

func GeneratePassword(passwordLength, minNum, minUpperCase int) string {
	const (
		lowerCharSet = "abcdedfghijklmnopqrst"
		upperCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numberSet    = "0123456789"
		allCharSet   = lowerCharSet + upperCharSet + numberSet
	)

	var password strings.Builder

	// Set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	// Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return string(inRune)
}

func GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)

	return string(hash), err
}

func CheckPasswordRequirements(password string) error {
	if len(password) < 6 {
		return errors.New("password length must exceed 6 or more")
	}

	return nil
}
