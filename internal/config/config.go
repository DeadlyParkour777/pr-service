package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTP_PORT       string
	DatabaseURL     string
	JWTSecret       string
	OpenAPISpecPath string
}

func NewConfig() (*Config, error) {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return nil, fmt.Errorf("DB_HOST environment variable is not set")
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		return nil, fmt.Errorf("DB_PORT environment variable is not set")
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return nil, fmt.Errorf("DB_USER environment variable is not set")
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is not set")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, fmt.Errorf("DB_NAME environment variable is not set")
	}
	dbURL := DatabaseConnString(dbUser, dbPassword, dbHost, dbPort, dbName)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	specPath := os.Getenv("OPENAPI_SPEC_PATH")
	if specPath == "" {
		specPath = "./docs/openapi.yml"
	}

	return &Config{
		HTTP_PORT:       port,
		DatabaseURL:     dbURL,
		JWTSecret:       jwtSecret,
		OpenAPISpecPath: specPath,
	}, nil
}

func DatabaseConnString(user, password, host, port, dbname string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname)
}
