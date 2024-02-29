package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"riza/app/core/events"
	"riza/app/core/hookers"
	"riza/app/deps"
	"riza/app/web/apiutils"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type GithubHandler struct{}

type GithubWebhookRequest struct {
	Data map[string]interface{}
}

func GithubWebhookLoggingHandler(c echo.Context) error {
	req := &GithubWebhookRequest{Data: make(map[string]interface{})}

	if err := c.Bind(&req.Data); err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			apiutils.ErrorResponse{
				ErrorCode:    apiutils.CodeWebhookRegistrationFailed,
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

func ValidateAndPublishWebhook(dep *deps.AppDeps) echo.HandlerFunc {
	// we dont need to send no error messages to github webhook servers.
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var repo string
		err := echo.PathParamsBinder(c).String("repo", &repo).BindError()
		if err != nil {
			logrus.WithError(err).Error("failed to get project in params")
			return c.JSON(http.StatusBadRequest, `{}`)
		}

		r := c.Request().Body
		defer r.Close()

		b, err := io.ReadAll(r)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to marshal request body")
			return c.JSON(http.StatusInternalServerError, `{}`)
		}

		svc := dep.HookRegistrationSvc
		hook, err := svc.FindByRepoProvider(ctx, &hookers.FindHookByProvider{
			Provider: hookers.GithubProvider,
			RepoPath: repo,
		})
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to find registered webhook")
			return c.JSON(http.StatusBadRequest, `{}`)
		}

		githubSignature := c.Request().Header.Get("x-hub-signature-256")
		if len(githubSignature) > 0 {
			logrus.WithContext(ctx).Infoln("github signature received in header")
		}

		hash := hmac.New(sha256.New, []byte(hook.Secrets))
		if _, err := hash.Write(b); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to compute hmac of request body")
			// return here
		}

		expectedHash := hex.EncodeToString(hash.Sum(nil))

		logrus.WithContext(ctx).Infoln("hash compare ", expectedHash, githubSignature)

		if ("sha256=" + expectedHash) != githubSignature {
			logrus.WithContext(ctx).WithError(err).Error("invalid payload signature.")
			// return here
		}

		ev, err := events.ToGithubEvents(b)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal github event.")
			logrus.WithContext(ctx).Debugln(string(b))

			return c.JSON(http.StatusBadRequest, `{}`)
		}

		// TODO: validate GithubEvents
		// TODO: Debounce

		err = dep.GithubEventsRepository.Create(ctx, hook.RepoPath, ev)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to push to queueue")
			return c.JSON(http.StatusTeapot, `{}`)
		}

		return c.JSON(http.StatusOK, `{}`)
	}
}
