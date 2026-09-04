// Package config assembles dependency DSNs from discrete environment values.
// Interpolating a password directly into a URL lets URL-significant characters
// such as @, :, / and # corrupt the address, so credentials are encoded here
// instead of by whichever tool composes the environment.
package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
)

// Getenv reads one environment value; the process environment is passed in so
// that assembly stays testable without mutating global state.
type Getenv func(string) string

func PostgresDSN(getenv Getenv) (string, error) {
	host, port, err := endpoint(getenv, "POSTGRES_HOST", "POSTGRES_PORT", "5432")
	if err != nil {
		return "", err
	}
	user, password := getenv("POSTGRES_USER"), getenv("POSTGRES_PASSWORD")
	if user == "" || password == "" {
		return "", fmt.Errorf("POSTGRES_USER and POSTGRES_PASSWORD are required")
	}
	database := getenv("POSTGRES_DB")
	if database == "" {
		return "", fmt.Errorf("POSTGRES_DB is required")
	}
	sslMode := getenv("POSTGRES_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + database,
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}
	return dsn.String(), nil
}

func RedisDSN(getenv Getenv) (string, error) {
	host, port, err := endpoint(getenv, "REDIS_HOST", "REDIS_PORT", "6379")
	if err != nil {
		return "", err
	}
	password := getenv("REDIS_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("REDIS_PASSWORD is required")
	}
	database := getenv("REDIS_DB")
	if database == "" {
		database = "0"
	}
	if _, err := strconv.ParseUint(database, 10, 8); err != nil {
		return "", fmt.Errorf("REDIS_DB must be a small non-negative integer")
	}
	dsn := url.URL{
		Scheme: "redis",
		User:   url.UserPassword("", password),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + database,
	}
	return dsn.String(), nil
}

func endpoint(getenv Getenv, hostKey, portKey, defaultPort string) (string, string, error) {
	host := getenv(hostKey)
	if host == "" {
		return "", "", fmt.Errorf("%s is required", hostKey)
	}
	port := getenv(portKey)
	if port == "" {
		port = defaultPort
	}
	number, err := strconv.ParseUint(port, 10, 16)
	if err != nil || number == 0 {
		return "", "", fmt.Errorf("%s must be a TCP port", portKey)
	}
	return host, port, nil
}
