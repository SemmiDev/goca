package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/logger"
)

func LoggerMiddleware(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Get request ID from Fiber's Locals (set by requestid middleware)
		requestID := GetRequestID(c)

		// Create a new context with the request ID
		ctx := context.WithValue(c.UserContext(), config.RequestIDContextKey, requestID)
		c.SetUserContext(ctx) // Update Fiber's context with the new one

		// Continue to next handler
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Convert query to string, default to empty string if nil
		query := ""
		if qs := c.Request().URI().QueryString(); qs != nil {
			query = string(qs)
		}

		// Safely get response body length
		var bytesSent int
		if body := c.Response().Body(); body != nil {
			bytesSent = len(body)
		}

		// Safely get user agent
		userAgent := c.Get("User-Agent")
		if userAgent == "" {
			userAgent = "unknown"
		}

		// Safely get IP
		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		// --- Pakai slice of key-value (approach 1) ---
		fields := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status_code", fmt.Sprintf("%d", statusCode),
			"duration_ms", fmt.Sprintf("%d", duration.Milliseconds()),
			"bytes_sent", fmt.Sprintf("%d", bytesSent),
			"user_agent", userAgent,
			"ip", ip,
			"query", query,
		}

		logEntry := log.WithContext(c.UserContext()).With(fields...)

		// Log based on status code
		if statusCode >= 500 {
			logEntry.Error("HTTP Request - Server Error")
		} else if statusCode >= 400 {
			logEntry.Warn("HTTP Request - Client Error")
		} else {
			logEntry.Info("HTTP Request - Success")
		}

		return err
	}
}
