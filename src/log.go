package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	appLog  *log.Logger
	authLog *log.Logger
)

func init() {
	appLog = createLogger("app.log")
	authLog = createLogger("auth.log")
}

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
