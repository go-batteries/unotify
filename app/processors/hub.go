package processors

import (
	"context"
	"unotify/app/consumers"
	"unotify/app/deps"
	"unotify/app/pkg/config"
	"unotify/app/pkg/exmachine"
	"unotify/app/pkg/workerpool"

	"github.com/sirupsen/logrus"
)

type Stager func(context.Context, *deps.AppDeps, *config.AppConfig, exmachine.StateReactorEngine)

func GetStagers(provisionerName string) Stager {
	logrus.Infoln("getting processor for provisioner", provisionerName)

	switch provisionerName {
	case "atlassian":
		return StageJiraProcessor
	default:
		return NoopProcessor
	}
}

func NoopProcessor(context.Context, *deps.AppDeps, *config.AppConfig, exmachine.StateReactorEngine) {
	panic("not implemented")
}

func StageJiraProcessor(ctx context.Context, dep *deps.AppDeps, cfg *config.AppConfig, reactor exmachine.StateReactorEngine) {
	jp, err := NewJiraProcessor(cfg, DefaultJiraEventChanSize, reactor)
	if err != nil {
		logrus.Fatal("exiting.... ", err)
	}

	wp := workerpool.NewWorkerPool(ctx, 4, jp.ProcessEach)
	wp.Start(ctx, jp.EventChannel)

	ghc := consumers.NewGithubEventConsumer(dep.GithubResqueue)
	go ghc.Start(ctx, "providers::github")

	consumers.GithubDispatcher(ctx, ghc.EventChannel, jp.EventChannel)
}
