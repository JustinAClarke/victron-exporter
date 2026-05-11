package main

import (
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// getEnv reads an environment variable value or returns the provided fallback.
// This simplifies configuration by allowing defaults to be supplied for missing values.
func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

// getIntEnv reads an integer environment variable or returns the fallback if the
// variable is missing. The application will fatal log if the variable exists but
// cannot be parsed as an integer, since this indicates misconfiguration.
func getIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			log.WithFields(log.Fields{
				"env_var":   key,
				"env_value": value}).
				WithError(err).Fatal("Unable to parse ENV VAR as an INT")
		}

		return i
	}
	log.WithFields(log.Fields{
		"env_var":        key,
		"fallback_value": fallback}).
		Debug("Unable to find ENV VAR, falling back to default value")

	return fallback
}

// getBoolEnv reads a boolean environment variable or returns the fallback if the
// variable is not present. Values such as 1, t, T, TRUE, true, True, 0, f, F,
// FALSE, false, False are supported by strconv.ParseBool.
func getBoolEnv(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(value)
		if err != nil {
			log.WithFields(log.Fields{
				"env_var":   key,
				"env_value": value}).
				WithError(err).Fatal("Unable to parse ENV VAR as a BOOL")
		}

		return b
	}
	log.WithFields(log.Fields{
		"env_var":        key,
		"fallback_value": fallback}).
		Debug("Unable to find ENV VAR, falling back to default value")

	return fallback
}

// getDurationEnv reads a duration environment variable or returns the fallback
// duration if the environment variable is missing. It fatals on invalid duration
// strings because an incorrect value would break the periodic poll loop.
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(value)
		if err != nil {
			log.WithFields(log.Fields{
				"env_var":   key,
				"env_value": value}).
				WithError(err).Fatal("Unable to parse ENV VAR as a DURATION")
		}

		return d
	}
	log.WithFields(log.Fields{
		"env_var":        key,
		"fallback_value": fallback}).
		Debug("Unable to find ENV VAR, falling back to default value")

	return fallback
}
