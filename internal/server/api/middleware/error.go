package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sammidev/goca/internal/pkg/logger"
	"github.com/sammidev/goca/internal/pkg/response"
)

func ErrorHandlerMiddleware(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		return response.HandleErrorAPI(c, err)
	}
}
