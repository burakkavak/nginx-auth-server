package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

// This file handles logic related to logging.
// The application produces two log files located at <executable_dir>/data/app.log and <executable_dir>/data/auth.log.

var (
	// appLog is the logger used for any application related logs (errors, warnings, informative, ...)
	appLog *log.Logger

	// authLog is the logger used for any authentication related logs (logins, logouts, ...)
	authLog *log.Logger
)

func init() {
	appLog = createLogger("app.log")
	authLog = createLogger("auth.log")
}

// createLogger creates and returns a log.Logger instance that logs to stdout and additionally the given filename.
// If the log files are not creatable/writable, a default logger (just stdout) is returned.
func createLogger(filename string) *log.Logger {
	logsDirectoryPath := fmt.Sprintf("%s/data", GetExecutableDirectory())

	if _, err := os.Stat(logsDirectoryPath); os.IsNotExist(err) {
		if err = os.MkdirAll(logsDirectoryPath, 0770); err != nil {
			log.Printf("cannot create data/logs directory at '%s'. application will log to stdout only. %s", logsDirectoryPath, err)
			return log.Default()
		}
	}

	logFilePath := fmt.Sprintf("%s/%s", logsDirectoryPath, filename)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		log.Printf("cannot create/write logfile at '%s'. application will log to stdout only. %s", logFilePath, err)
		return log.Default()
	}

	return log.New(io.MultiWriter(os.Stdout, logFile), "", log.Lshortfile|log.LstdFlags)
}
