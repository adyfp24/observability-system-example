package rest

import (
	"github.com/go-playground/validator/v10"
	"strconv"

	"github.com/adyfp24/okejek-go-service/order-service/internal/modules/order/core/dto"
	"github.com/adyfp24/okejek-go-service/order-service/internal/modules/order/core/usecase"
	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	uc *usecase.UseCase
}

func NewController(uc *usecase.UseCase) *Controller {
	return &Controller{uc: uc}
}

func (ctrl *Controller) RegisterRoutes(router fiber.Router) {
	router.Post("/orders", ctrl.Create)
	router.Get("/orders/:id", ctrl.Get)
	router.Patch("/orders/:id/status", ctrl.UpdateStatus)
}

func (ctrl *Controller) Create(c *fiber.Ctx) error {
	var req dto.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	res, err := ctrl.uc.CreateOrder(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (ctrl *Controller) Get(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID order tidak valid",
		})
	}

	res, err := ctrl.uc.GetOrder(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (ctrl *Controller) UpdateStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID order tidak valid",
		})
	}

	var req struct {
		Status string `json:"status" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	err = ctrl.uc.UpdateOrderStatus(c.UserContext(), id, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Status order berhasil diperbarui",
	})
}
