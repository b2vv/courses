package http

import (
	"fmt"
	"strconv"
)

// Config carries the admin-stats HTTP layer's environment-derived settings.
type Config struct {
	SingleInstructor bool
}

// NewConfig builds a config from its parts.
func NewConfig(singleInstructor bool) Config {
	return Config{SingleInstructor: singleInstructor}
}

// ConfigFromEnvironment reads PLATFORM_SINGLE_INSTRUCTOR, defaulting to true
// when unset (the platform runs with a single instructor by default).
func ConfigFromEnvironment(getenv func(string) string) (Config, error) {
	singleInstructor, err := boolWithDefault(getenv("PLATFORM_SINGLE_INSTRUCTOR"), true, "PLATFORM_SINGLE_INSTRUCTOR")
	if err != nil {
		return Config{}, err
	}
	return NewConfig(singleInstructor), nil
}

// boolWithDefault returns the fallback for an unset value and a named error for
// a value that is neither true nor false.
func boolWithDefault(value string, fallback bool, name string) (bool, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", name)
	}
	return parsed, nil
}
