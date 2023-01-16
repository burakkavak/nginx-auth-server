package main

import (
	"os"
	"path/filepath"
)

// This file contains any helper functions that can be useful in any file/module.

// CheckFileReadable returns nil if the given file path is readable.
// If the file is not readable, the function will return an error.
func CheckFileReadable(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	_ = file.Close()

	return err
}

// GetExecutableDirectory returns the parent directory path of the currently executed binary.
// Refer to https://stackoverflow.com/a/18537419
func GetExecutableDirectory() string {
	ex, err := os.Executable()

	if err != nil {
		appLog.Fatalf("fatal error: could not get parent directory of the executed binary. %s\n", err)
	}

	exPath := filepath.Dir(ex)

	return exPath
}
