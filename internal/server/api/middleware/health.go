package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthCheckMiddleware handles health check requests
func HealthCheckMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/health" || c.Path() == "/healthz" {
			return c.JSON(map[string]interface{}{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"uptime":    time.Since(time.Now()).Seconds(), // This would be calculated properly
			})
		}

		return c.Next()
	}
}
