package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"riza/app/web/webhook"

	"github.com/go-batteries/diaper"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	Port        int
	DatabaseURL string
	Env         string
}

var LogLevelMap = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"error": logrus.ErrorLevel,
	"_":     logrus.InfoLevel,
}

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "dev"
	}

	dc := diaper.DiaperConfig{
		Providers:      diaper.Providers{diaper.EnvProvider{}},
		DefaultEnvFile: "app.env",
	}

	cfgMap, err := dc.ReadFromFile(env, "./config/")
	if err != nil {
		logrus.WithError(err).Fatal("failed to load config from .env")
	}

	cfg := AppConfig{
		Port:        cfgMap.MustGetInt("port"),
		DatabaseURL: cfgMap.MustGet("database_url").(string),
		Env:         env,
	}

	logLevelStr := cfgMap.MustGet("log_level").(string)
	if logLevelStr == "" {
		logLevelStr = "_"
	}

	logrus.SetLevel(LogLevelMap[logLevelStr])

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	group := e.Group("/arij")

	group.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	group.POST("/webhook/payload", webhook.GithubWebhookHandler)

	logrus.Infof("starting port at %d\n", cfg.Port)
	data, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		logrus.WithError(err).Debugln("failed to get routes")
	} else {
		logrus.Debugln(string(data))
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.Port)))
}
