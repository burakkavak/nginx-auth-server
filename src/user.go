package main

import (
	"encoding/json"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
)

type User struct {
	Username  string `json:"username"`
	Password  string `json:"password"`  // Password :: salted & hashed
	OtpSecret []byte `json:"otpSecret"` // OtpSecret :: encrypted OTP secret key
}

func GetUsers() []User {
	db := initDatabase()
	defer db.Close()

	var users []User

	_ = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))

		if bucket == nil {
			return nil
		}

		_ = bucket.ForEach(func(key, value []byte) error {
			user := User{}
			_ = json.Unmarshal(value, &user)
			users = append(users, user)

			return nil
		})

		return nil
	})

	return users
}

func CreateUser(user *User) error {
	if GetUserByUsername(user.Username) != nil {
		return errors.New("user with username '" + user.Username + "' already exists")
	}

	if len(GetCookiesByUsername(user.Username)) != 0 {
		return errors.New(fmt.Sprintf("error: username '%s' already has cookies saved in database. this username is possibly linked to a LDAP user.", user.Username))
	}

	db := initDatabase()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("users"))
		bucket := tx.Bucket([]byte("users"))

		buffer, err := json.Marshal(user)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return bucket.Put([]byte(user.Username), buffer)
	})
}

func RemoveUser(username string) error {
	if GetUserByUsername(username) == nil {
		return errors.New("user with username '" + username + "' does not exist")
	}

	db := initDatabase()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("users"))
		bucket := tx.Bucket([]byte("users"))

		return bucket.Delete([]byte(username))
	})
}

// GetUserByUsername Looks up username in database and returns the user if found. Returns nil if the user was not found.
func GetUserByUsername(username string) *User {
	db := initDatabase()
	defer db.Close()

	var user *User

	_ = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))

		if bucket == nil {
			return nil
		}

		v := bucket.Get([]byte(username))

		if v == nil {
			return nil
		}

		_ = json.Unmarshal(v, &user)

		return nil
	})

	return user
}

func GetUserByUsernameCaseInsensitive(username string) *User {
	users := GetUsers()

	for _, user := range users {
		if user.Username == username {
			return &user
		}
	}

	return nil
}
