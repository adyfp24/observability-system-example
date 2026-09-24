package rest

import (
	"github.com/go-playground/validator/v10"
	"strconv"

	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/modules/user/core/dto"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/modules/user/core/usecase"
	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	uc *usecase.UseCase
}

func NewController(uc *usecase.UseCase) *Controller {
	return &Controller{uc: uc}
}

func (ctrl *Controller) RegisterRoutes(router fiber.Router) {
	router.Post("/auth/register", ctrl.Register)
	router.Post("/auth/login", ctrl.Login)
	router.Get("/users/:id", ctrl.GetProfile)
}

func (ctrl *Controller) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	// Anda bisa memanggil validator di sini jika diperlukan
	err := ctrl.uc.Register(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Registrasi berhasil",
	})
}

func (ctrl *Controller) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	res, err := ctrl.uc.Login(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (ctrl *Controller) GetProfile(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	res, err := ctrl.uc.GetProfile(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
