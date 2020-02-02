package main

import (
	"fmt"
	"github.com/google/logger"
	"log"
	"os"
)

func initLogger(logPath string) {
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

	if err != nil {
		printFatal("Could not open log file: %s\n", err)
	}

	logger.Init("channel-bot", true, true, file)
	logger.SetFlags(log.LstdFlags)

	logger.Info("Initialized logger")
}

func printFatal(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	os.Exit(1)
}
