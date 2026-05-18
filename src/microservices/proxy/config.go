package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port                 int
	MonolithURL          string
	MoviesServiceURL     string
	EventsServiceURL     string
	GradualMigration     bool
	MoviesMigrationPercent int
}

func LoadConfig() (*Config, error) {
	port, err := getEnvInt("PORT", 8000)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	monolithURL := getEnv("MONOLITH_URL", "http://localhost:8080")
	moviesServiceURL := getEnv("MOVIES_SERVICE_URL", "http://localhost:8081")
	eventsServiceURL := getEnv("EVENTS_SERVICE_URL", "http://localhost:8082")

	gradualMigration := getEnvBool("GRADUAL_MIGRATION", false)
	moviesMigrationPercent, err := getEnvInt("MOVIES_MIGRATION_PERCENT", 0)
	if err != nil {
		return nil, fmt.Errorf("invalid MOVIES_MIGRATION_PERCENT: %w", err)
	}

	return &Config{
		Port:                 port,
		MonolithURL:          monolithURL,
		MoviesServiceURL:     moviesServiceURL,
		EventsServiceURL:     eventsServiceURL,
		GradualMigration:     gradualMigration,
		MoviesMigrationPercent: moviesMigrationPercent,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return n, nil
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
