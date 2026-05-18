package main

import (
	"fmt"
	"log"

	"github.com/cinemaabyss/microservices/proxy/api"
	"github.com/labstack/echo/v4"
	middleware "github.com/oapi-codegen/echo-middleware"
)
func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	proxy := api.NewProxy(
		cfg.MonolithURL,
		cfg.MoviesServiceURL,
		cfg.EventsServiceURL,
		cfg.GradualMigration,
		cfg.MoviesMigrationPercent,
	)

	e := echo.New()

	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatalf("failed to load OpenAPI spec: %v", err)
	}
	swagger.Servers = nil

	e.Use(middleware.OapiRequestValidator(swagger))

 strictHandler := api.NewStrictHandler(proxy, []api.StrictMiddlewareFunc{})

	api.RegisterHandlers(e, strictHandler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting Strangler Fig Proxy on %s", addr)
	log.Printf("  Monolith:            %s", cfg.MonolithURL)
	log.Printf("  Movies Service:      %s", cfg.MoviesServiceURL)
	log.Printf("  Events Service:      %s", cfg.EventsServiceURL)
	log.Printf("  Gradual Migration:   %v", cfg.GradualMigration)
	log.Printf("  Movies Migration %%:  %d", cfg.MoviesMigrationPercent)

	log.Fatal(e.Start(addr))
}
