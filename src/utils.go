package main

import (
	"os"
)

// CheckFileReadable :: returns nil if the given file path is readable
func CheckFileReadable(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	_ = file.Close()

	return err
}
