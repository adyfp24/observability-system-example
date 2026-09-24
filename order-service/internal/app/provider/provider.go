package provider

import (
	"github.com/adyfp24/okejek-go-service/order-service/internal/app/config"
	"github.com/adyfp24/okejek-go-service/order-service/internal/app/helper"
	"github.com/adyfp24/okejek-go-service/order-service/internal/modules/order"
	"github.com/redis/go-redis/v9"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Provider struct {
	App       *fiber.App
	Config    *config.Config
	DB        *gorm.DB
	Validator *config.Validator
	Logger    *helper.Logger
	Cache     *redis.Client
}

func (p *Provider) Provide() {
	order.Provide(p.App, p.DB, p.Validator, p.Cache, p.Config)
}
