package main

import (
	"context"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
	cfg := LoadConfig()

	producer := NewKafkaProducer(cfg.KafkaBrokers)
	defer producer.Close()

	consumer := NewKafkaConsumer(cfg.KafkaBrokers)
	defer consumer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go consumer.Start(ctx)

	handler := NewEventHandler(producer)

	e := echo.New()
	e.GET("/api/events/health", handler.Health)
	e.POST("/api/events/movie", handler.CreateMovieEvent)
	e.POST("/api/events/user", handler.CreateUserEvent)
	e.POST("/api/events/payment", handler.CreatePaymentEvent)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting events service on %s", addr)
	log.Fatal(e.Start(addr))
}
