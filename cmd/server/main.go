package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	"unotify/app/deps"
	"unotify/app/pkg/config"
	"unotify/app/web/webhook"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	var appPort string
	flag.StringVar(&appPort, "p", "9091", "application port overrides env")
	flag.Parse()

	cfg := config.BuildAppConfig(env)

	config.SetupLogger(cfg.Env, cfg.LogLevel)

	dep := deps.BuildAppDeps(cfg)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	group := e.Group("/conduit")

	// group.POST("/webhook/payload", webhook.GithubWebhookLoggingHandler)

	group.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	group.GET("/version", func(c echo.Context) error {
		return c.String(http.StatusOK, "3")
	})

	group.POST("/webhooks/github/:repo/payload", webhook.ValidateAndPublishWebhook(dep))

	// remove this from load balancer
	// to put it back, do group.PATCH
	e.PATCH("/webhooks/update", webhook.RegisterWebHook(dep.HookRegistrationSvc, true))

	group.POST("/webhooks/register", webhook.RegisterWebHook(dep.HookRegistrationSvc, false))
	group.GET("/webhooks/list", webhook.ListRegisteredHooksForProvider(dep.HookRegistrationSvc))
	group.GET("/webhooks/find", webhook.FindRegisteredHooks(dep.HookRegistrationSvc))

	// group.GET("/oauthapp/jira/callback", )

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
		logrus.Println("server serving at port ", appPort)

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
