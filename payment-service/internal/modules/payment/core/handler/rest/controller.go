package rest

import (
	"github.com/go-playground/validator/v10"
	"os"
	"strconv"
	"time"

	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/dto"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/usecase"
	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	uc *usecase.UseCase
}

func NewController(uc *usecase.UseCase) *Controller {
	return &Controller{uc: uc}
}

func (ctrl *Controller) RegisterRoutes(router fiber.Router) {
	router.Get("/wallets/:user_id", ctrl.GetWallet)
	router.Post("/wallets/topup", ctrl.TopUp)
	router.Post("/wallets/pay", ctrl.Pay)
	router.Post("/wallets/absence", ctrl.Absence)
}

func (ctrl *Controller) GetWallet(c *fiber.Ctx) error {
	userIDStr := c.Params("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID tidak valid",
		})
	}

	res, err := ctrl.uc.GetWallet(c.UserContext(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (ctrl *Controller) TopUp(c *fiber.Ctx) error {
	var req dto.TopUpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	err := ctrl.uc.TopUp(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Top-up saldo berhasil diproses",
	})
}

func (ctrl *Controller) Pay(c *fiber.Ctx) error {
	if os.Getenv("DEMO_MODE") == "true" {
		ms, _ := strconv.Atoi(c.Get("X-Demo-Delay-Ms"))
		if ms > 0 && ms <= 2000 {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}
	var req dto.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	err := ctrl.uc.Pay(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Pembayaran berhasil diproses",
	})
}

func (ctrl *Controller) Absence(c *fiber.Ctx) error {
	var req dto.AbsenceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	err := ctrl.uc.ProcessDriverAbsence(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Potongan absensi harian driver berhasil",
	})
}
