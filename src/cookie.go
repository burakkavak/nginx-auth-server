package main

import (
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"time"
)

// Cookie :: refer to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`   // Value :: salted & hashed
	Expires  time.Time `json:"expires"` // example: 'Wed, 21 Oct 2015 07:28:00 GMT'
	Domain   string    `json:"domain"`
	Username string    `json:"username"`
	HttpOnly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
}

func SaveCookie(cookie Cookie) error {
	encodedCookieHash := GenerateHash(cookie.Value)

	cookie.Value = encodedCookieHash

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

// GetCookiesByUsername Looks up cookies specific to a user in database and returns the cookies.
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

// GetCookieByValue Looks up cookie in database and returns the cookie if found. Returns nil if the cookie was not found.
// TODO: function performance is horrible, matching the plain cookie value to all argon hashes in the database takes too long
func GetCookieByValue(cookieValue string) *Cookie {
	cookies := GetCookies()

	return func() *Cookie {
		for _, cookie := range cookies {
			if CompareHashAndPassword(cookie.Value, cookieValue) == nil {
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

func DeleteCookieByValue(cookieValue string) error {
	cookie := GetCookieByValue(cookieValue)

	if cookie == nil {
		return errors.New("error: cannot find cookie in database")
	} else {
		return DeleteCookie(cookie)
	}
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
	cookie := GetCookieFromCache(cookieValue)

	if cookie == nil {
		cookie = GetCookieByValue(cookieValue)
	}

	if cookie == nil {
		return errors.New("error: cookie not found")
	} else if cookie.Expires.Before(time.Now()) {
		DeleteCookie(cookie)
		DeleteCookieFromCache(cookie)
		return errors.New("error: cookie is expired and was deleted")
	} else {
		SaveCookieToCache(cookie, cookieValue)
		return nil
	}
}
