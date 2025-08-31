package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sammidev/goca/internal/pkg/apperror"
)

func ValidateJSONMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentType := c.Get("Content-Type")
			if contentType != "" && contentType != "modules/json" {
				return apperror.NewAppError(
					apperror.ErrCodeInvalidInput,
					"Content-Type must be modules/json",
				)
			}

			// Check if body is empty for methods that require it
			if len(c.Body()) == 0 {
				return apperror.NewAppError(
					apperror.ErrCodeInvalidInput,
					"Request body is required",
				)
			}
		}

		return c.Next()
	}
}
