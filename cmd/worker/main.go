package main

import (
	"context"
	"os"
	"os/signal"
	"unotify/app/consumers"
	"unotify/app/deps"
	"unotify/app/pkg/config"
	"unotify/app/pkg/workerpool"
	"unotify/app/processors"

	"github.com/sirupsen/logrus"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.BuildAppConfig(env)
	config.SetupLogger(cfg.Env, cfg.LogLevel)

	dep := deps.BuildAppDeps(cfg)

	jp, err := processors.NewJiraProcessor(cfg, processors.DefaultJiraEventChanSize)
	if err != nil {
		logrus.Fatal("exiting.... ", err)
	}

	wp := workerpool.NewWorkerPool(ctx, 4, jp.ProcessEach)
	wp.Start(ctx, jp.EventChannel)

	ghc := consumers.NewGithubEventConsumer(dep.GithubResqueue)
	go ghc.Start(ctx, "providers::github")

	consumers.GithubDispatcher(ctx, ghc.EventChannel, jp.EventChannel)
	<-ctx.Done()

	logrus.Infoln("worker stopped")
}
