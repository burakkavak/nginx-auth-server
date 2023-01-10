package main

import (
	"log"
	"os"
	"path/filepath"
)

// CheckFileReadable :: returns nil if the given file path is readable
func CheckFileReadable(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	_ = file.Close()

	return err
}

// GetExecutableDirectory :: returns the parent directory path of the currently executed binary (https://stackoverflow.com/a/18537419)
func GetExecutableDirectory() string {
	ex, err := os.Executable()

	if err != nil {
		log.Fatalf("fatal error: could not get parent directory of the executed binary. %s", err)
	}

	exPath := filepath.Dir(ex)

	return exPath
}
