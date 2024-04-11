package schema

import "fmt"

// HttpError 请求错误
type HttpError struct {
	Code   int
	ErrMsg string
}

func (h *HttpError) Error() string {
	return fmt.Sprintf("http status: %d, errMsg: %s", h.Code, h.ErrMsg)
}

func NewHttpError(code int, errMsg string) error {
	return &HttpError{code, errMsg}
}
