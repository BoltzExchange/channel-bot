package utils

import "github.com/google/logger"

func CheckError(service string, err error) {
	if err != nil {
		logger.Fatal("Could not initialize "+service+": ", err)
	}
}
