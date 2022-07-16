package main

import (
	bolt "go.etcd.io/bbolt"
	"log"
)

func initDatabase() *bolt.DB {
	db, err := bolt.Open("nginx-auth-server.db", 0600, nil)

	if err != nil {
		log.Fatal("could not open/create database. check working directory permissions")
	}

	return db
}
