package utils

import (
	"log"
	"os"

	"github.com/google/logger"
)

func InitLogger(logPath string) {
	var file *os.File

	if logPath != "" {
		var err error
		file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

		if err != nil {
			PrintFatal("Could not open log file: %s", err)
		}
	}

	logger.Init("channel-bot", true, false, file)
	logger.SetFlags(log.LstdFlags)

	logger.Info("Initialized logger")
}
