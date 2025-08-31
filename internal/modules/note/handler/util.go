package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/pkg/apperror"
)

func ParseUUIDParam(c *fiber.Ctx, param string) (uuid.UUID, error) {
	idStr := c.Params(param)
	if idStr == "" {
		return uuid.Nil, apperror.NewAppError(apperror.ErrCodeBadRequest, fmt.Sprintf("missing param %s", param))
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, apperror.NewAppError(apperror.ErrCodeBadRequest, fmt.Sprintf("invalid %s format", param))
	}

	return id, nil
}
