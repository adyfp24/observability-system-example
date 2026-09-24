package payment

import (
	"github.com/adyfp24/okejek-go-service/payment-service/internal/app/config"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/handler/rest"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/repository"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/router"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func Provide(app *fiber.App, db *gorm.DB, validator *config.Validator, cache *redis.Client, cfg *config.Config) {
	repo := repository.NewRepository(db, cache)
	uc := usecase.NewUseCase(repo, validator, cfg)
	ctrl := rest.NewController(uc)

	router.Route(app, ctrl)
}
