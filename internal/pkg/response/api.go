package response

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/apperror"
	"github.com/sammidev/goca/internal/pkg/request"
)

func HandleSuccessAPI(c *fiber.Ctx, statusCode int, message string, data any, paging *request.Paging) error {
	opts := []ResponseOption{}

	if id := c.Locals(config.RequestIDContextKey); id != nil {
		opts = append(opts, WithRequestID(id.(string)))
	}

	if paging != nil {
		opts = append(opts, WithPaging(paging))
	}

	successResponse := NewSuccessResponse(message, data, opts...)
	return c.Status(statusCode).JSON(successResponse)
}

func HandleErrorAPI(c *fiber.Ctx, err error) error {
	opts := []ResponseOption{}

	if id := c.Locals(config.RequestIDContextKey); id != nil {
		opts = append(opts, WithRequestID(id.(string)))
	}

	var (
		appErr        *apperror.AppError
		validationErr *apperror.ValidationErrors
		fiberErr      *fiber.Error
	)

	// Siapkan nilai default
	statusCode := http.StatusInternalServerError
	responseErrorCode := apperror.ErrCodeInternalError
	responseErrorMessage := "An internal error occurred"

	if errors.As(err, &validationErr) {
		statusCode = validationErr.HTTPStatusCode()
		responseErrorCode = validationErr.AppError.Code
		responseErrorMessage = validationErr.AppError.Message
		opts = append(opts, WithErrorDetails(validationErr.Fields.ToMap()))
	}

	if errors.As(err, &fiberErr) {
		err = apperror.NewBodyParserError()
	}

	if errors.As(err, &appErr) {
		statusCode = appErr.HTTPStatusCode()
		responseErrorCode = appErr.Code
		responseErrorMessage = appErr.Message
	}

	resp := NewErrorResponse(responseErrorMessage, responseErrorCode.String(), opts...)
	return c.Status(statusCode).JSON(resp)
}
