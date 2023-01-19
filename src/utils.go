package main

import (
	"embed"
	"github.com/gin-gonic/gin"
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

// GetFilenamesFromFS retrieves all files from the given path
// of the given embedded filesystem and returns the filenames as an array.
func GetFilenamesFromFS(fs embed.FS, path string) []string {
	files, err := fs.ReadDir(path)

	if err != nil {
		appLog.Fatalf("fatal error: could not read embedded css/js files. %s\n", err)
	}

	var filenames []string

	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	return filenames
}

// GetClientIpFromContext retrieves and returns the real client IP from the given Gin context
// using the 'X-Original-Remote-Addr' header set by NGINX.
func GetClientIpFromContext(c *gin.Context) string {
	clientIp := c.GetHeader("X-Original-Remote-Addr")

	if clientIp == "" {
		appLog.Print("warning: could not determine (real) client IP address - 'X-Original-Remote-Addr' header was not set by nginx")
	}

	return clientIp
}
