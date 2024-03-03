package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
	"unotify/app/consumers"
	"unotify/app/deps"
	"unotify/app/pkg/config"
	"unotify/app/pkg/exmachine"
	"unotify/app/pkg/workerpool"
	"unotify/app/processors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	var hclConfigDir string
	flag.StringVar(&hclConfigDir, "hcl-dir", "./config/statemachines", "directory for hcl config")

	var appPort string
	flag.StringVar(&appPort, "p", ":9093", "application port overrides env")
	flag.Parse()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	group := e.Group("/conduit-reactor")

	// group.POST("/webhook/payload", webhook.GithubWebhookLoggingHandler)

	group.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.BuildAppConfig(env)
	config.SetupLogger(cfg.Env, cfg.LogLevel)

	dep := deps.BuildAppDeps(cfg)

	filePath, err := filepath.Abs(filepath.Join(hclConfigDir, "jira.hcl"))
	if err != nil {
		logrus.WithError(err).Fatal("failed to load statmachine dir from", hclConfigDir)
	}

	reader := exmachine.HCLFileReader{}
	statenames := []string{"soc", "devhop"}
	states := []*exmachine.StateMachine{}

	for _, statename := range statenames {
		statemach, err := exmachine.Provision(
			ctx,
			statename,
			reader,
			filePath,
		)
		if err != nil {
			logrus.WithError(err).Fatal("failed to provision for machine", statename, "at", filePath)
		}

		states = append(states, statemach)
	}

	reactor := exmachine.BuildStateMachine(states...)

	jp, err := processors.NewJiraProcessor(cfg, processors.DefaultJiraEventChanSize, reactor)
	if err != nil {
		logrus.Fatal("exiting.... ", err)
	}

	wp := workerpool.NewWorkerPool(ctx, 4, jp.ProcessEach)
	wp.Start(ctx, jp.EventChannel)

	ghc := consumers.NewGithubEventConsumer(dep.GithubResqueue)
	go ghc.Start(ctx, "providers::github")

	consumers.GithubDispatcher(ctx, ghc.EventChannel, jp.EventChannel)

	go func() {
		logrus.Println("worker serving at port ", appPort)

		if err := e.Start(appPort); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()

	logrus.Infoln("worker stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
