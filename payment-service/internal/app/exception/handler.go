package exception

import (
	"errors"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/app/model"

	"github.com/gofiber/fiber/v2"
)

func Handler(ctx *fiber.Ctx, err error) error {
	var (
		badRequestError     *BadRequestError
		unauthorizedError   *UnauthorizedError
		forbiddenError      *ForbiddenError
		notFoundError       *NotFoundError
		internalServerError *InternalServerError
		notImplementedError *NotImplementedError
	)

	switch {
	case errors.As(err, &badRequestError):
		return ctx.Status(fiber.StatusBadRequest).JSON(model.Response{
			Code:    fiber.StatusBadRequest,
			Status:  "failed",
			Message: "Bad Request",
			Data:    badRequestError.Error(),
		})
	case errors.As(err, &unauthorizedError):
		return ctx.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Code:    fiber.StatusUnauthorized,
			Status:  "failed",
			Message: "Unauthorized",
			Data:    unauthorizedError.Error(),
		})
	case errors.As(err, &forbiddenError):
		return ctx.Status(fiber.StatusForbidden).JSON(model.Response{
			Code:    fiber.StatusForbidden,
			Status:  "failed",
			Message: "Forbidden",
			Data:    forbiddenError.Error(),
		})
	case errors.As(err, &notFoundError):
		return ctx.Status(fiber.StatusNotFound).JSON(model.Response{
			Code:    fiber.StatusNotFound,
			Status:  "failed",
			Message: "Not Found",
			Data:    notFoundError.Error(),
		})
	case errors.As(err, &internalServerError):
		return ctx.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Code:    fiber.StatusInternalServerError,
			Status:  "failed",
			Message: "Internal Server Error",
			Data:    internalServerError.Error(),
		})
	case errors.As(err, &notImplementedError):
		return ctx.Status(fiber.StatusNotImplemented).JSON(model.Response{
			Code:    fiber.StatusNotImplemented,
			Status:  "failed",
			Message: "Not Implemented",
			Data:    notImplementedError.Error(),
		})
	default:
		return ctx.Status(fiber.StatusInternalServerError).JSON(model.Response{
			Code:    fiber.StatusInternalServerError,
			Status:  "failed",
			Message: "Internal Server Error",
			Data:    err.Error(),
		})
	}
}
