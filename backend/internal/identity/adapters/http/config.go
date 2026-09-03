package http

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const SessionCookieName = "courseforge_session"

type Config struct {
	publicOrigin string
	secureCookie bool
}

func ConfigFromEnvironment(getenv func(string) string) (Config, error) {
	secure, err := strconv.ParseBool(getenv("COOKIE_SECURE"))
	if err != nil {
		return Config{}, fmt.Errorf("COOKIE_SECURE must be true or false")
	}
	return NewConfig(getenv("APP_ENV"), getenv("PUBLIC_ORIGIN"), secure)
}

func NewConfig(environment, origin string, secure bool) (Config, error) {
	if environment != "development" && environment != "test" && environment != "production" {
		return Config{}, fmt.Errorf("APP_ENV must be development, test or production")
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" && parsed.Path != "/" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return Config{}, fmt.Errorf("PUBLIC_ORIGIN must be an absolute HTTP origin")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return Config{}, fmt.Errorf("PUBLIC_ORIGIN must use HTTP or HTTPS")
	}
	hostname := strings.ToLower(parsed.Hostname())
	loopback := hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
	if !secure && !loopback {
		return Config{}, fmt.Errorf("insecure cookies require a loopback origin")
	}
	if environment == "production" && (!secure || parsed.Scheme != "https") {
		return Config{}, fmt.Errorf("production requires HTTPS and secure cookies")
	}
	return Config{publicOrigin: canonicalOrigin(parsed), secureCookie: secure}, nil
}

func canonicalOrigin(origin *url.URL) string {
	host := strings.ToLower(origin.Hostname())
	port := origin.Port()
	if port == "" {
		if origin.Scheme == "http" {
			port = "80"
		} else {
			port = "443"
		}
	}
	return strings.ToLower(origin.Scheme) + "://" + net.JoinHostPort(host, port)
}

func (config Config) acceptsOrigin(value string) bool {
	origin, err := url.Parse(value)
	return err == nil && origin.Scheme != "" && origin.Host != "" && origin.User == nil && origin.Path == "" && origin.RawQuery == "" && origin.Fragment == "" && canonicalOrigin(origin) == config.publicOrigin
}
