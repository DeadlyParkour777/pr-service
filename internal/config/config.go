package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTP_PORT   string
	DatabaseURL string
	JWTSecret   string
}

func NewConfig() (*Config, error) {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE URL environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	return &Config{
		HTTP_PORT:   port,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
	}, nil
}
