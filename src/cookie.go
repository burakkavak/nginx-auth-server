package main

import (
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"regexp"
	"time"
)

// Cookie :: refer to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`   // Value :: argon2 hash
	Expires  time.Time `json:"expires"` // example: 'Wed, 21 Oct 2015 07:28:00 GMT'
	Domain   string    `json:"domain"`
	Username string    `json:"username"`
	HttpOnly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
}

// SaveCookie saves a cookie to the database.
// Returns nil if the cookie was saved successfully.
func SaveCookie(cookie Cookie) error {
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

// GetCookies fetches all cookies from the database and returns them.
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

// GetCookiesByUsername looks up cookies specific to a user in database and returns the cookies.
func GetCookiesByUsername(username string) []Cookie {
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

			if cookie.Username == username {
				cookies = append(cookies, cookie)
			}

			return nil
		})

		return nil
	})

	return cookies
}

// GetCookieByValue looks up plaintext cookie value and the corresponding user in the database
// and returns the cookie if found. This function will try to match the given plaintext cookie value
// to the argon2 hash saved in the database, therefore it is a time-intensive function.
// Returns nil if the cookie was not found.
func GetCookieByValue(cookieValue string, username string) *Cookie {
	var cookies []Cookie

	cookies = GetCookiesByUsername(username)

	for _, cookie := range cookies {
		if CompareHashAndPassword(cookie.Value, cookieValue) == nil {
			return &cookie
		}
	}

	return nil
}

// PurgeCookies deletes all cookies in the database.
func PurgeCookies() error {
	db := initDatabase()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("cookies"))
	})
}

// DeleteCookie deletes a specific Cookie from the database.
func DeleteCookie(cookie *Cookie) error {
	if cookie == nil {
		return errors.New("error: provided cookie is nil")
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

// DeleteCookiesByUsername deletes all cookies specific to the given username.
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

// VerifyCookie returns the Cookie and nil if the given token is valid.
// Example for token param: '$username=foo,$value=kC6......LOh'.
// Returns nil and an error if the cookie was not found or expired.
func VerifyCookie(token string) (*Cookie, error) {
	username, cookieValue, err := DecodeAuthToken(token)

	if err != nil {
		return nil, err
	}

	cookie := GetCookieFromCache(cookieValue)

	if cookie == nil {
		cookie = GetCookieByValue(cookieValue, username)
	}

	if cookie == nil {
		return nil, errors.New("error: cookie not found")
	} else if cookie.Expires.Before(time.Now()) {
		err := DeleteCookie(cookie)

		if err != nil {
			return nil, errors.New("error: could not delete expired cookie from database")
		} else {
			DeleteCookieFromCache(cookie)
			return nil, errors.New("error: cookie is expired and was deleted")
		}
	} else {
		SaveCookieToCache(cookie, cookieValue)
		return cookie, nil
	}
}

// DecodeAuthToken decodes the given token and returns the username and plain cookie value.
// Returns an error if the given token did not match the expected syntax.
// Example for token param: '$username=foo,$value=kC6......LOh'.
func DecodeAuthToken(token string) (username string, cookieValue string, err error) {
	// match the username and cookie value from the new syntax ($username=<username>,$value=<value>)
	// filtering the cookies by username before matching the plain cookie value to the argon hash
	// in the database improves performance
	regex := regexp.MustCompile(`\$username=(?P<username>.+?),\$value=(?P<value>.+)`)

	if regex.MatchString(token) {
		matches := regex.FindStringSubmatch(token)

		username = matches[regex.SubexpIndex("username")]
		cookieValue = matches[regex.SubexpIndex("value")]

		return username, cookieValue, nil
	} else {
		return "", "", errors.New("auth token does not match syntax")
	}
}
