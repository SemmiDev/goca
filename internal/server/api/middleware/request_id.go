package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/config"
)

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return requestid.New(requestid.Config{
		Header:     string(config.RequestIDHeaderKey),
		Generator:  func() string { return uuid.Must(uuid.NewV7()).String() },
		ContextKey: config.RequestIDContextKey,
	})
}

func GetRequestID(c *fiber.Ctx) string {
	if id := c.Locals(config.RequestIDContextKey); id != nil {
		if reqID, ok := id.(string); ok && reqID != "" {
			return reqID
		}
	}
	return "unknown"
}
