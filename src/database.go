package main

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"os"
)

var databaseFilePath string

// init will check if the database file is readable/creatable by the application.
// This function will panic if the database could not be read/created for some reason.
func init() {
	databaseFilePath = fmt.Sprintf("%s/nginx-auth-server.db", GetExecutableDirectory())

	if _, err := os.Stat(databaseFilePath); os.IsNotExist(err) {
		// use new database file path in 'data'-subfolder if a database does not exist at the old path
		databaseDirectoryPath := fmt.Sprintf("%s/data", GetExecutableDirectory())

		if _, err := os.Stat(databaseDirectoryPath); os.IsNotExist(err) {
			if err = os.MkdirAll(databaseDirectoryPath, 0770); err != nil {
				appLog.Fatalf("fatal error: cannot create data directory at '%s'. %s", databaseDirectoryPath, err)
			}
		}

		databaseFilePath = fmt.Sprintf("%s/nginx-auth-server.db", databaseDirectoryPath)
	}
}

// initDatabase opens and returns the bbolt database.
// This function will panic if the database could not be accessed for some reason.
func initDatabase() *bolt.DB {
	// TODO: can cause delays on concurrent calls
	db, err := bolt.Open(databaseFilePath, 0660, nil)

	if err != nil {
		appLog.Fatalf("could not open/create database. %s\n", err)
	}

	return db
}
