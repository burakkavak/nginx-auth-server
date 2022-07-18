package main

import (
	"encoding/json"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

// todo: periodically remove expired cookies

// Cookie :: refer to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`   // Value :: salted & hashed
	Expires  time.Time `json:"expires"` // example: 'Wed, 21 Oct 2015 07:28:00 GMT'
	Domain   string    `json:"domain"`
	Username string    `json:"username"`
}

func SaveCookie(cookie Cookie) error {
	if GetUserByUsername(cookie.Username) == nil {
		return errors.New(fmt.Sprintf("user with username '%s' does not exist", cookie.Username))
	}

	saltedCookieValue, err := GenerateHash(cookie.Value)

	if err != nil {
		log.Fatalf("error while salting the cookie value for database storage: %s", err)
	}

	cookie.Value = saltedCookieValue

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

func GetCookies() []Cookie {
	db := initDatabase()
	defer db.Close()

	var cookies []Cookie

	_ = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("cookies"))

		if bucket == nil {
			return nil
		}

		_ = bucket.ForEach(func(key, value []byte) error {
			cookie := Cookie{}
			_ = json.Unmarshal(value, &cookie)
			cookies = append(cookies, cookie)

			return nil
		})

		return nil
	})

	return cookies
}

// GetCookieByValue Looks up cookie in database and returns the cookie if found. Returns nil if the cookie was not found.
func GetCookieByValue(cookieValue string) *Cookie {
	cookies := GetCookies()

	return func() *Cookie {
		for _, cookie := range cookies {
			if bcrypt.CompareHashAndPassword([]byte(cookie.Value), []byte(cookieValue)) == nil {
				return &cookie
			}
		}

		return nil
	}()
}

func PurgeCookies() error {
	db := initDatabase()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("cookies"))
	})
}

func DeleteCookie(cookieValue string) error {
	cookie := GetCookieByValue(cookieValue)

	if cookie == nil {
		return errors.New("error: cannot find cookie in database")
	}

	db := initDatabase()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("cookies"))

		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(cookie.Value))
	})
}

// DeleteCookiesByUsername :: Delete all cookies for a given username
func DeleteCookiesByUsername(username string) error {
	db := initDatabase()
	defer db.Close()

	return db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("cookies"))

		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(key, value []byte) error {
			cookie := Cookie{}
			_ = json.Unmarshal(value, &cookie)

			if cookie.Username == username {
				return bucket.Delete([]byte(cookie.Value))
			} else {
				return nil
			}
		})
	})
}

// VerifyCookie :: Returns nil if the cookie is valid
func VerifyCookie(cookieValue string) error {
	cookie := GetCookieByValue(cookieValue)

	if cookie == nil || cookie.Expires.Before(time.Now()) {
		return errors.New("error: cookie not found or expired")
	} else {
		return nil
	}
}
