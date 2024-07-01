package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"unotify/app/core/events"
	"unotify/app/core/hookers"
	"unotify/app/deps"
	"unotify/app/pkg/ds"
	"unotify/app/web/apiutils"

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

// TODO: Make configurable
var AllowedProviders = ds.ToSet("github")

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

		var provider string
		err = echo.PathParamsBinder(c).String("provider", &provider).BindError()
		if err != nil {
			logrus.WithError(err).Error("failed to get project in params")
			return c.JSON(http.StatusBadRequest, `{}`)
		}

		if !AllowedProviders.Has(provider) {
			logrus.Error("unregistered provider", provider)
			return c.JSON(http.StatusForbidden, `{}`)
		}

		r := c.Request().Body
		defer r.Close()

		b, err := io.ReadAll(r)
		if err != nil {
			logrus.
				WithContext(ctx).
				WithError(err).
				Error("failed to marshal request body")

			return c.JSON(http.StatusInternalServerError, `{}`)
		}

		// provider = hookers.GithubProvider
		svc := dep.HookRegistrationSvc
		hook, err := svc.FindByRepoProvider(ctx, &hookers.FindHookByProvider{
			Provider: provider,
			RepoID:   repo,
		})
		if err != nil {
			logrus.
				WithContext(ctx).
				WithError(err).
				Error("failed to find registered webhook")

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
			logrus.WithContext(ctx).Error("invalid payload signature.")
			// return here
			return c.JSON(http.StatusBadRequest, `{}`)
		}

		ev, err := events.ScoopGithubEvents(b)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal github event.")
			logrus.WithContext(ctx).Debugln(string(b))

			return c.JSON(http.StatusBadRequest, `{}`)
		}

		// TODO: validate GithubEvents
		// TODO: Debounce

		// The initial idea was to Enqueue to a list, per project
		// But rn, it just does per hook.Provider basic. You can
		// Turn on that sharding, by uncommenting key creation in EnqueMsg
		err = dep.GithubEventsRepository.Create(ctx, "events::github", ev)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to push to queueue")
			return c.JSON(http.StatusTeapot, `{}`)
		}

		return c.JSON(http.StatusOK, `{}`)
	}
}
