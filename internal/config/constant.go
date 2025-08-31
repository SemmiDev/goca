package config

import "time"

const (
	RequestIDHeaderKey  string = "X-Request-ID"
	RequestIDContextKey string = "requestid"
)

const (
	EmailConfirmationExpInMin = 15 * time.Minute
	OTPCodeLength             = 6
)

const (
	AuthRateLimiterKey = "auth"
)
