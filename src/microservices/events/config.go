package main

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	KafkaBrokers string
}

func LoadConfig() *Config {
	portStr := os.Getenv("PORT")
	port := 8082
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	return &Config{
		Port:         port,
		KafkaBrokers: brokers,
	}
}
