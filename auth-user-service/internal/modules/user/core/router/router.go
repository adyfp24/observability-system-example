package router

import (
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/modules/user/core/handler/rest"
	"github.com/gofiber/fiber/v2"
)

func Route(app *fiber.App, ctrl *rest.Controller) {
	api := app.Group("/api/v1")
	ctrl.RegisterRoutes(api)
}
