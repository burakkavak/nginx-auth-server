package main

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
)

func initDatabase() *bolt.DB {
	db, err := bolt.Open(fmt.Sprintf("%s/nginx-auth-server.db", GetExecutableDirectory()), 0600, nil)

	if err != nil {
		log.Fatal("could not open/create database. check working directory permissions")
	}

	return db
}
