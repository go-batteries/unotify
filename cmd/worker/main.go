package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"unotify/app/consumers"
	"unotify/app/deps"
	"unotify/app/pkg/config"
	"unotify/app/pkg/exmachine"
	"unotify/app/pkg/workerpool"
	"unotify/app/processors"

	"github.com/sirupsen/logrus"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	var hclConfigDir string
	flag.StringVar(&hclConfigDir, "hcl-dir", "./config/statemachines", "directory for hcl config")
	flag.Parse()

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

	// soc, err := exmachine.Provision(
	// 	ctx,
	// 	"soc",
	// 	reader,
	// 	filePath,
	// )

	// devops, err := exmachine.Provision(
	// 	ctx,
	// 	"devhops",
	// 	reader,
	// 	filePath,
	// )
	// if err != nil {
	// 	logrus.WithError(err).Fatal("failed to provision devops statemachine", filePath)
	// }

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
	<-ctx.Done()

	logrus.Infoln("worker stopped")
}
