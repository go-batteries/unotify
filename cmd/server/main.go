package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"riza/app/deps"
	"riza/app/pkg/config"
	"riza/app/web/webhook"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

var LogLevelMap = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"error": logrus.ErrorLevel,
	"":      logrus.InfoLevel,
}

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "dev"
	}

	var appPort string
	flag.StringVar(&appPort, "p", "9091", "application port overrides env")
	flag.Parse()

	cfg := config.BuildAppConfig(env)
	dep := deps.BuildAppDeps(cfg)

	logrus.SetLevel(LogLevelMap[cfg.LogLevel])

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	group := e.Group("/arij")

	group.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	group.GET("/version", func(c echo.Context) error {
		return c.String(http.StatusOK, "2")
	})

	group.POST("/webhook/payload", webhook.GithubWebhookLoggingHandler)

	group.POST("/webhook/github/:project/payload", webhook.GithubWebhookHandler)

	group.POST("/webhooks/register", webhook.RegisterWebHook(dep.HookRegistrationSvc))
	group.GET("/webhooks/find", webhook.FindRegisteredHooks(dep.HookRegistrationSvc))
	group.GET("/webhooks/query", webhook.ListRegisteredHooksForProvider(dep.HookRegistrationSvc))

	if appPort == "" {
		appPort = fmt.Sprintf(":%d", cfg.Port)
	}

	if !strings.HasPrefix(appPort, ":") {
		appPort = ":" + appPort
	}

	logrus.Infoln("starting port at ", appPort)
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		logrus.WithError(err).Debugln("failed to get routes")
	} else {
		logrus.Debugln(string(data))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
	go func() {
		if err := e.Start(appPort); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
