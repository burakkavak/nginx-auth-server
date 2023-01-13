package main

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
)

func initDatabase() *bolt.DB {
	db, err := bolt.Open(fmt.Sprintf("%s/nginx-auth-server.db", GetExecutableDirectory()), 0660, nil)

	if err != nil {
		appLog.Fatalf("could not open/create database. %s\n", err)
	}

	return db
}
