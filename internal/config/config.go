package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppHost           string
	AppPort           string
	WSReadBufferSize  int
	WSWriteBufferSize int
	WatcherPoll       time.Duration
	TLSEnabled        bool
	TLSCertFile       string
	TLSKeyFile        string
	DB                DBGroup
}

type DBGroup struct {
	Authorization DBConfig
	Supply        DBConfig
	Reference     DBConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func Load(path string) (Config, error) {
	if err := loadDotEnv(path); err != nil {
		return Config{}, err
	}

	watcherPoll, err := time.ParseDuration(getEnv("WATCHER_POLL_INTERVAL", "1s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse WATCHER_POLL_INTERVAL: %w", err)
	}

	cfg := Config{
		AppHost:           getEnv("APP_HOST", "0.0.0.0"),
		AppPort:           getEnv("APP_PORT", "8080"),
		WSReadBufferSize:  getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
		WSWriteBufferSize: getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
		WatcherPoll:       watcherPoll,
		TLSEnabled:        getEnvAsBool("TLS_ENABLED", false),
		TLSCertFile:       getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:        getEnv("TLS_KEY_FILE", ""),
		DB: DBGroup{
			Authorization: loadDBConfig("AUTH"),
			Supply:        loadDBConfig("SUPPLY"),
			Reference:     loadDBConfig("REFERENCE"),
		},
	}

	return cfg, nil
}

func loadDBConfig(prefix string) DBConfig {
	return DBConfig{
		Host:     getEnv(prefix+"_DB_HOST", "127.0.0.1"),
		Port:     getEnvAsInt(prefix+"_DB_PORT", 3306),
		User:     getEnv(prefix+"_DB_USER", "root"),
		Password: getEnv(prefix+"_DB_PASSWORD", ""),
		Name:     getEnv(prefix+"_DB_NAME", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open dotenv: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan dotenv: %w", err)
	}

	return nil
}
