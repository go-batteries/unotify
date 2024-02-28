package webhook

import (
	"encoding/json"
	"net/http"
	"riza/app/web/apiutils"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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

	b, err := json.MarshalIndent(req.Data, " ", " ")
	if err != nil {
		logrus.WithError(err).Error("failed to marshal json response")
	} else {
		logrus.Println("response ", string(b))
	}

	return c.JSON(
		http.StatusOK,
		apiutils.SuccessResponse{
			Success: true,
			Data:    req.Data,
		},
	)
}
