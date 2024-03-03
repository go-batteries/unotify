package config

import (
	"github.com/sirupsen/logrus"
)

// For time being we will configure logger here

func SetupLogger(env string, logLevel string) {
	LogLevelMap := map[string]logrus.Level{
		"debug": logrus.DebugLevel,
		"error": logrus.ErrorLevel,
		"":      logrus.InfoLevel,
	}

	if logLevel != "debug" && logLevel != "error" {
		logLevel = ""
	}

	logrus.SetLevel(LogLevelMap[logLevel])
	logrus.SetFormatter(&logrus.JSONFormatter{})

	if env == "local" {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	logrus.SetReportCaller(true)
}
