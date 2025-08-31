package response

import "github.com/sammidev/goca/internal/pkg/request"

type Response struct {
	Success bool         `json:"success" example:"true"`
	Message string       `json:"message" example:"Success"`
	Data    any          `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
	Meta    *Meta        `json:"meta,omitempty"`
}

type Meta struct {
	Paging    *request.Paging `json:"paging,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
}

type ErrorDetail struct {
	Code    string            `json:"code,omitempty" example:"error_code"`
	Message string            `json:"message" example:"Error message"`
	Details map[string]string `json:"details,omitempty"`
}

func NewSuccessResponse(message string, data any, options ...ResponseOption) *Response {
	res := &Response{
		Success: true,
		Message: message,
		Data:    data,
	}

	for _, opt := range options {
		opt(res)
	}

	return res
}

func NewErrorResponse(message string, code string, options ...ResponseOption) Response {
	res := Response{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: code, Message: message},
	}

	for _, opt := range options {
		opt(&res)
	}

	return res
}

type ResponseOption func(*Response)

func WithPaging(paging *request.Paging) ResponseOption {
	return func(r *Response) {
		if r.Meta == nil {
			r.Meta = &Meta{}
		}
		r.Meta.Paging = paging
	}
}

func WithRequestID(id string) ResponseOption {
	return func(r *Response) {
		if r.Meta == nil {
			r.Meta = &Meta{}
		}
		r.Meta.RequestID = id
	}
}

func WithErrorDetails(details map[string]string) ResponseOption {
	return func(r *Response) {
		if r.Error != nil {
			r.Error.Details = details
		}
	}
}
