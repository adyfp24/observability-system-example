package main

import (
	"context"
	"fmt"
	"github.com/adyfp24/okejek-go-service/order-service/internal/app/telemetry"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/adyfp24/okejek-go-service/order-service/internal/app/config"
	"github.com/adyfp24/okejek-go-service/order-service/internal/app/http"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	shutdown := telemetry.Init("order-service")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdown(ctx)
	}()
	var (
		cfg       = config.NewConfig()
		db        = config.InitDB(cfg)
		validator = config.NewValidator()
		cache     = config.InitCache(cfg)
	)

	app := http.NewFiber(cfg, db, validator, cache)

	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		app.Shutdown()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() { <-ctx.Done(); app.ShutdownWithTimeout(5 * time.Second) }()
	port := cfg.GetInt("APP_PORT", 3001)
	log.Printf("Server order_service starting on port %d", port)

	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
