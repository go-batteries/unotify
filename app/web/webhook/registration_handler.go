package webhook

import (
	"net/http"
	"unotify/app/core/hookers"
	"unotify/app/web/apiutils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func FindRegisteredHooks(svc *hookers.HookerService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		req := &hookers.FindHookByProvider{}

		if err := c.Bind(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to bind request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRegistrationFailed,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to validate request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRequestParamMissing,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		res, err := svc.FindByRepoProvider(ctx, req)
		if err != nil {
			return c.JSON(
				http.StatusNotFound,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeHookerMissing,
					ErrorMessage: apiutils.ErrHooksMissing.Error(),
				},
			)
		}

		return c.JSON(
			http.StatusOK,
			apiutils.SuccessResponse{
				Success: true,
				Data:    res,
			},
		)
	}
}

func ListRegisteredHooksForProvider(svc *hookers.HookerService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		req := &hookers.FindHookByProvider{}

		if err := c.Bind(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to bind request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRegistrationFailed,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to validate request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRequestParamMissing,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		res, err := svc.List(ctx, req)
		if err != nil {
			return c.JSON(
				http.StatusNotFound,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeHookerMissing,
					ErrorMessage: apiutils.ErrHooksMissing.Error(),
				},
			)
		}

		return c.JSON(
			http.StatusOK,
			apiutils.SuccessResponse{
				Success: true,
				Data:    res,
			},
		)
	}
}

func RegisterWebHook(svc *hookers.HookerService, forceUpdate bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := &hookers.RegisterHookRequest{ForceUpdate: forceUpdate}
		ctx := c.Request().Context()

		if err := c.Bind(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to bind request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRegistrationFailed,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to validate request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRegistrationFailed,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		findhook := &hookers.FindHookByProvider{
			RepoID:   req.RepoID,
			Provider: req.Provider,
		}

		_, err := svc.FindByRepoProvider(ctx, findhook)
		if err == nil {
			return c.JSON(
				http.StatusConflict,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeHookExists,
					ErrorMessage: apiutils.ErrDuplicateRegistration.Error(),
				},
			)
		}

		resp, err := svc.Register(ctx, req)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to register webhook to db")

			return c.JSON(
				http.StatusInternalServerError,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeInternalServerError,
					ErrorMessage: apiutils.ErrInternalServerError.Error(),
				},
			)
		}

		return c.JSON(
			http.StatusCreated,
			apiutils.SuccessResponse{
				Success: true,
				Data:    resp,
			},
		)
	}
}

func ImportWebHook(svc *hookers.HookerService) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := &hookers.ImportHookRequest{}
		ctx := c.Request().Context()

		if err := c.Bind(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to bind request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRegistrationFailed,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to validate request")

			return c.JSON(
				http.StatusBadRequest,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeWebhookRegistrationFailed,
					ErrorMessage: apiutils.ErrInvalidRequest.Error(),
				},
			)
		}

		resp, err := svc.Import(ctx, req)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to import webhook to db")

			return c.JSON(
				http.StatusInternalServerError,
				apiutils.ErrorResponse{
					ErrorCode:    apiutils.CodeInternalServerError,
					ErrorMessage: apiutils.ErrInternalServerError.Error(),
				},
			)
		}

		return c.JSON(
			http.StatusCreated,
			apiutils.SuccessResponse{
				Success: true,
				Data:    resp,
			},
		)
	}
}
