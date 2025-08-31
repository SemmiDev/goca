package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sammidev/goca/internal/pkg/logger"
)

func RecoverMiddleware(log logger.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.WithContext(c.UserContext()).Error("Panic recovered",
				"error", e,
				"method", c.Method(),
				"path", c.Path(),
				"ip", c.IP(),
			)
		},
	})
}
