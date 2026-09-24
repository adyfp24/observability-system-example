package http

import (
	"bytes"
	"context"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/app/config"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/app/exception"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/app/provider"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/app/telemetry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/common/expfmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func NewFiber(cfg *config.Config, db *gorm.DB, validator *config.Validator, cache *redis.Client) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: exception.Handler})
	app.Get("/metrics", func(c *fiber.Ctx) error {
		families, err := telemetry.Registry.Gather()
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		var body bytes.Buffer
		encoder := expfmt.NewEncoder(&body, expfmt.FmtText)
		for _, family := range families {
			if err := encoder.Encode(family); err != nil {
				return c.Status(500).SendString(err.Error())
			}
		}
		c.Set(fiber.HeaderContentType, string(expfmt.FmtText))
		return c.Send(body.Bytes())
	})
	app.Get("/health", func(c *fiber.Ctx) error {
		sql, err := db.DB()
		if err != nil || sql.PingContext(context.Background()) != nil {
			return c.SendStatus(503)
		}
		return c.SendString("ok")
	})
	app.Use(telemetry.Middleware())
	app.Use(recover.New())
	p := provider.Provider{App: app, Config: cfg, DB: db, Validator: validator, Cache: cache}
	p.Provide()
	return app
}
