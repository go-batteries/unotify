package webhook

import (
	"net/http"
	"riza/app/web/apiutils"

	"github.com/labstack/echo/v4"
)

type GithubHandler struct{}

type GithubWebhookRequest struct {
	Data map[string]interface{}
}

func GithubWebhookHandler(c echo.Context) error {
	req := &GithubWebhookRequest{Data: make(map[string]interface{})}

	if err := c.Bind(&req.Data); err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			apiutils.ErrorResponse{
				ErrorCode:    apiutils.GithubInternalServerErrorCode,
				ErrorMessage: apiutils.ErrInternalServerError.Error(),
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		apiutils.SuccessResponse{
			Success: true,
			Data:    req.Data,
		},
	)
}
