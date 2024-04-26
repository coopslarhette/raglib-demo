package api

import (
	"github.com/go-chi/render"
	"net/http"
)

type ErrResponse struct {
	Error          error  `json:"-"`
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	AppCode        int64  `json:"code,omitempty"`
	ErrorText      string `json:"error,omitempty"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrorMalformedRequest(err error) render.Renderer {
	return &ErrResponse{
		Error:          err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Malformed request.",
		ErrorText:      err.Error(),
	}
}

func InternalServerError(err error) render.Renderer {
	return &ErrResponse{
		Error:          err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal server error.",
		ErrorText:      err.Error(),
	}
}
