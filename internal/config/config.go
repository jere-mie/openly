package config

import (
	"bufio"
	"os"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	AdminPassword string
	Port          string
	Host          string
	DatabasePath  string
}

// Load reads configuration from environment variables (and .env file).
func Load() *Config {
	loadEnvFile(".env")

	cfg := &Config{
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin"),
		Port:          getEnv("PORT", "8080"),
		Host:          getEnv("HOST", "localhost"),
		DatabasePath:  getEnv("DATABASE_PATH", "openly.db"),
	}

	return cfg
}

// loadEnvFile reads a .env file and sets environment variables (without overriding existing ones).
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
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
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
