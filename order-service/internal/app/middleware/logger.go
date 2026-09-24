package middleware

import (
	"encoding/json"
	"time"

	"github.com/adyfp24/okejek-go-service/order-service/internal/app/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

func LoggingMiddleware(logger *helper.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/health" {
			return c.Next()
		}

		requestBody := c.Body()
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Milliseconds()
		responseBody := c.Response().Body()

		ctx := helper.FiberContextToContext(c)

		var reqBody, resBody map[string]interface{}
		if err := json.Unmarshal(requestBody, &reqBody); err != nil {
			reqBody = map[string]interface{}{"raw": string(requestBody)}
		}
		if err := json.Unmarshal(responseBody, &resBody); err != nil {
			resBody = map[string]interface{}{"raw": string(responseBody)}
		}

		reqHeaders := convertHeadersFiber(c.GetReqHeaders())
		resHeaders := make(map[string]interface{})
		c.Response().Header.VisitAll(func(key, value []byte) {
			resHeaders[string(key)] = string(value)
		})

		logger.LogAPIRequest(
			ctx,
			c.Path(),
			c.Method(),
			c.Response().StatusCode(),
			duration,
			reqHeaders,
			reqBody,
			resHeaders,
			resBody,
		)

		return err
	}
}

func convertHeadersFiber(headers map[string][]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range headers {
		if len(v) == 1 {
			result[k] = v[0]
		} else {
			result[k] = v
		}
	}
	return result
}
