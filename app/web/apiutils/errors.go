package apiutils

import "errors"

// error codes
const (
	apierrorCode = iota + 100

	GithubInternalServerErrorCode
)

// errors
var (
	ErrInternalServerError = errors.New("internal_server_error")
)

type ErrorResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_msg"`
	Success      bool   `json:"success"`
}
