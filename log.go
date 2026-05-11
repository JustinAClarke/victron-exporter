package main

import (
	log "github.com/sirupsen/logrus"
)

// setLogLevel maps an integer configuration value to the logrus logging level.
// This allows the application to be configured with a simple numeric environment
// variable or flag instead of a logrus-specific level string.
func setLogLevel(logLevel int) {
	switch logLevel {
	case 0:
		log.SetLevel(log.DebugLevel)
	case 1:
		log.SetLevel(log.InfoLevel)
	case 2:
		log.SetLevel(log.WarnLevel)
	case 3:
		log.SetLevel(log.ErrorLevel)
	case 4:
		log.SetLevel(log.FatalLevel)
	}
}
