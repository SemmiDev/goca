package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sammidev/goca/internal/pkg/apperror"
	"github.com/sammidev/goca/internal/pkg/response"
	"github.com/sammidev/goca/internal/pkg/token"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "user"
)

var (
	ErrInvalidAuthFormat = errors.New("format token otentikasi tidak valid atau tidak didukung")
	ErrEmptyAuthHeader   = errors.New("header Authorization tidak disediakan")
	ErrInvalidAuthToken  = errors.New("token otentikasi tidak valid")
	ErrInvalidToken      = errors.New("token tidak valid atau telah kedaluwarsa")
)

func AuthMiddleware(tokenMaker token.Token) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authHeader := ctx.Get(authorizationHeaderKey)
		if authHeader == "" {
			appErr := apperror.NewAppError(apperror.ErrCodeUnauthorized, "Authorization header tidak ditemukan.")
			return response.HandleErrorAPI(ctx, appErr)
		}

		fields := strings.Fields(authHeader)
		if len(fields) != 2 || strings.ToLower(fields[0]) != authorizationTypeBearer {
			appErr := apperror.NewAppError(apperror.ErrCodeUnauthorized, "Format token tidak sesuai.")
			return response.HandleErrorAPI(ctx, appErr)
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			appErr := apperror.NewAppError(apperror.ErrCodeUnauthorized, "Token tidak valid atau telah kadaluwarsa.")
			return response.HandleErrorAPI(ctx, appErr)
		}

		ctx.Locals(authorizationPayloadKey, payload)
		return ctx.Next()
	}
}

func GetUser(ctx *fiber.Ctx) *token.Payload {
	payload, ok := ctx.Locals(authorizationPayloadKey).(*token.Payload)
	if !ok {
		return nil
	}
	return payload
}
