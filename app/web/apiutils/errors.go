package apiutils

import "errors"

// error codes
const (
	_ = iota

	CodeInternalServerError
)

const (
	_ = iota + 200

	CodeWebhookRegistrationFailed
	CodeWebhookRequestParamMissing
	CodeHookerMissing
	CodeHookExists
)

const (
	_ = iota + 300

	CodeGithubInternalServerError
)

// errors
var (
	ErrInternalServerError   = errors.New("internal_server_error")
	ErrInvalidRequest        = errors.New("malformed_request")
	ErrHooksMissing          = errors.New("unregistered_webhook")
	ErrDuplicateRegistration = errors.New("duplicate_registration")
)

type ErrorResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_msg"`
	Success      bool   `json:"success"`
}
