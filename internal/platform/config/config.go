package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Host              string
	Port              string
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	SessionCookieName string
	SessionTTL        time.Duration
	SessionSecure     bool
	BcryptCost        int
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Host:              getenv("HOST", "0.0.0.0"),
		Port:              getenv("PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		RedisAddr:         redisAddr(),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		SessionCookieName: getenv("SESSION_COOKIE_NAME", "session_id"),
		SessionTTL:        30 * 24 * time.Hour,
		BcryptCost:        bcrypt.DefaultCost,
		ReadHeaderTimeout: 10 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	var err error
	if cfg.SessionSecure, err = getbool("SESSION_COOKIE_SECURE", false); err != nil {
		return Config{}, err
	}
	if cfg.RedisDB, err = getint("REDIS_DB", 0); err != nil {
		return Config{}, err
	}
	if cfg.BcryptCost, err = getint("BCRYPT_COST", cfg.BcryptCost); err != nil {
		return Config{}, err
	}
	if err := validateBcryptCost(cfg.BcryptCost); err != nil {
		return Config{}, err
	}
	if cfg.SessionTTL, err = getduration("SESSION_TTL", cfg.SessionTTL); err != nil {
		return Config{}, err
	}
	if cfg.ReadHeaderTimeout, err = getduration("READ_HEADER_TIMEOUT", cfg.ReadHeaderTimeout); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = getduration("SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func redisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}

	host := getenv("REDIS_HOST", "localhost")
	port := getenv("REDIS_PORT", "6379")
	return net.JoinHostPort(host, port)
}

func getbool(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return parsed, nil
}

func getint(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func validateBcryptCost(cost int) error {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return fmt.Errorf("BCRYPT_COST must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}
	return nil
}

func getduration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return parsed, nil
}
