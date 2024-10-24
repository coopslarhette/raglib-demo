package api

import (
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type ErrorCode int

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeMalformedRequest
	ErrCodeInternalServer
)

type ErrResponse struct {
	HTTPStatusCode int       `json:"-"`
	Code           ErrorCode `json:"code"`
	Message        string    `json:"message"`
	Details        string    `json:"details,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	e.Log()
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func NewErrorResponse(httpStatus int, code ErrorCode, message string, details string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: httpStatus,
		Code:           code,
		Message:        message,
		Details:        details,
	}
}

func MalformedRequest(details string) render.Renderer {
	return NewErrorResponse(
		http.StatusBadRequest,
		ErrCodeMalformedRequest,
		"Malformed request",
		details,
	)
}

func InternalServerError(details string) render.Renderer {
	return NewErrorResponse(
		http.StatusInternalServerError,
		ErrCodeInternalServer,
		"Internal server error",
		details,
	)
}

// Log logs the full error details for internal use
func (e *ErrResponse) Log() {
	slog.Error("error type HTTP response returned", "code", e.Code, "message", e.Message, "details", e.Details)
}
