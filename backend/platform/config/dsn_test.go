package config_test

import (
	"testing"

	"courses/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	redisclient "github.com/redis/go-redis/v9"
)

// A password an operator might reasonably choose, carrying every character that
// changes how a URL parses.
const hostilePassword = "p@ss:w/rd?#[]&= x"

func environment(values map[string]string) config.Getenv {
	return func(key string) string { return values[key] }
}

func TestPostgresDSNSurvivesURLSignificantPasswordCharacters(t *testing.T) {
	t.Parallel()
	dsn, err := config.PostgresDSN(environment(map[string]string{
		"POSTGRES_HOST":     "postgres",
		"POSTGRES_DB":       "courses",
		"POSTGRES_USER":     "courses",
		"POSTGRES_PASSWORD": hostilePassword,
	}))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("pgx rejected the assembled DSN: %v", err)
	}
	if parsed.ConnConfig.Password != hostilePassword {
		t.Fatalf("password=%q want %q", parsed.ConnConfig.Password, hostilePassword)
	}
	if parsed.ConnConfig.Host != "postgres" || parsed.ConnConfig.Port != 5432 || parsed.ConnConfig.Database != "courses" || parsed.ConnConfig.User != "courses" {
		t.Fatalf("endpoint=%s:%d database=%s user=%s", parsed.ConnConfig.Host, parsed.ConnConfig.Port, parsed.ConnConfig.Database, parsed.ConnConfig.User)
	}
}

func TestRedisDSNSurvivesURLSignificantPasswordCharacters(t *testing.T) {
	t.Parallel()
	dsn, err := config.RedisDSN(environment(map[string]string{
		"REDIS_HOST":     "redis",
		"REDIS_PASSWORD": hostilePassword,
	}))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := redisclient.ParseURL(dsn)
	if err != nil {
		t.Fatalf("go-redis rejected the assembled DSN: %v", err)
	}
	if parsed.Password != hostilePassword {
		t.Fatalf("password=%q want %q", parsed.Password, hostilePassword)
	}
	if parsed.Addr != "redis:6379" || parsed.DB != 0 {
		t.Fatalf("addr=%s db=%d", parsed.Addr, parsed.DB)
	}
}

func TestDSNAssemblyRejectsIncompleteOrInvalidEnvironments(t *testing.T) {
	t.Parallel()
	complete := map[string]string{
		"POSTGRES_HOST": "postgres", "POSTGRES_DB": "courses", "POSTGRES_USER": "courses", "POSTGRES_PASSWORD": "secret",
		"REDIS_HOST": "redis", "REDIS_PASSWORD": "secret",
	}
	for name, override := range map[string]map[string]string{
		"missing postgres host":       {"POSTGRES_HOST": ""},
		"missing postgres user":       {"POSTGRES_USER": ""},
		"missing postgres password":   {"POSTGRES_PASSWORD": ""},
		"missing postgres database":   {"POSTGRES_DB": ""},
		"postgres port not a port":    {"POSTGRES_PORT": "not-a-port"},
		"postgres port zero":          {"POSTGRES_PORT": "0"},
		"postgres port too large":     {"POSTGRES_PORT": "70000"},
		"missing redis host":          {"REDIS_HOST": ""},
		"missing redis password":      {"REDIS_PASSWORD": ""},
		"redis database not a number": {"REDIS_DB": "one"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			values := map[string]string{}
			for key, value := range complete {
				values[key] = value
			}
			for key, value := range override {
				values[key] = value
			}
			getenv := environment(values)
			_, postgresErr := config.PostgresDSN(getenv)
			_, redisErr := config.RedisDSN(getenv)
			if postgresErr == nil && redisErr == nil {
				t.Fatal("invalid environment produced two usable DSNs")
			}
		})
	}
}
