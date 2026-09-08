package http

import (
	"fmt"
	"strconv"
)

const (
	defaultHeartbeatSeconds    = 30
	defaultOwnershipTTLSeconds = 180
)

type Config struct {
	heartbeatSeconds    int
	ownershipTTLSeconds int
}

func ConfigFromEnvironment(getenv func(string) string) (Config, error) {
	heartbeatSeconds, err := positiveSeconds(getenv("PLAYBACK_HEARTBEAT_SECONDS"), defaultHeartbeatSeconds, "PLAYBACK_HEARTBEAT_SECONDS")
	if err != nil {
		return Config{}, err
	}
	ownershipTTLSeconds, err := positiveSeconds(getenv("PLAYBACK_OWNERSHIP_TTL_SECONDS"), defaultOwnershipTTLSeconds, "PLAYBACK_OWNERSHIP_TTL_SECONDS")
	if err != nil {
		return Config{}, err
	}
	return NewConfig(heartbeatSeconds, ownershipTTLSeconds)
}

func NewConfig(heartbeatSeconds, ownershipTTLSeconds int) (Config, error) {
	if heartbeatSeconds <= 0 || ownershipTTLSeconds <= 0 {
		return Config{}, fmt.Errorf("playback intervals must be positive")
	}
	if heartbeatSeconds > ownershipTTLSeconds/3 {
		return Config{}, fmt.Errorf("playback requires heartbeatSeconds * 3 <= ownershipTtlSeconds")
	}
	return Config{heartbeatSeconds: heartbeatSeconds, ownershipTTLSeconds: ownershipTTLSeconds}, nil
}

func (config Config) HeartbeatSeconds() int { return config.heartbeatSeconds }

func (config Config) OwnershipTTLSeconds() int { return config.ownershipTTLSeconds }

func positiveSeconds(value string, fallback int, name string) (int, error) {
	if value == "" {
		return fallback, nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive decimal integer", name)
	}
	return seconds, nil
}
