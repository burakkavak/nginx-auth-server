package main

import (
	"encoding/json"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"math/rand"
	"strings"
	"time"
)

// Cookie :: refer to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Expires  time.Time `json:"expires"` // example: 'Wed, 21 Oct 2015 07:28:00 GMT'
	Domain   string    `json:"domain"`
	Username string    `json:"username"`
}

func GenerateAuthCookie(username string) Cookie {
	const (
		lowerCharSet = "abcdedfghijklmnopqrst"
		upperCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numberSet    = "0123456789"
		allCharSet   = lowerCharSet + upperCharSet + numberSet
	)

	var (
		value        strings.Builder
		length       = 72
		minNum       = 10
		minUpperCase = 30
	)

	// Set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		value.WriteString(string(numberSet[random]))
	}

	// Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		value.WriteString(string(upperCharSet[random]))
	}

	remainingLength := length - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		value.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(value.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return Cookie{
		Name:     "Auth",
		Value:    string(inRune),
		Expires:  time.Now().AddDate(0, 0, 7),
		Domain:   GetDomain(),
		Username: username,
	}
}

func CreateCookie(cookie *Cookie) error {
	if GetUserByUsername(cookie.Username) == nil {
		return errors.New(fmt.Sprintf("user with username '%s' does not exist", cookie.Username))
	}

	db := initDatabase()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("cookies"))
		bucket := tx.Bucket([]byte("cookies"))

		buffer, err := json.Marshal(cookie)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return bucket.Put([]byte(cookie.Value), buffer)
	})

}
