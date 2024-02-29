package config

import (
	"github.com/go-batteries/diaper"
	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	Port        int
	LogLevel    string
	DatabaseURL string
	RedisURL    string
	Env         string
}

func BuildAppConfig(env string) *AppConfig {
	dc := diaper.DiaperConfig{
		Providers:      diaper.Providers{diaper.EnvProvider{}},
		DefaultEnvFile: "app.env",
	}

	cfgMap, err := dc.ReadFromFile(env, "./config/")
	if err != nil {
		logrus.WithError(err).Fatal("failed to load config from .env")
	}

	cfg := &AppConfig{
		Port:     cfgMap.MustGetInt("port"),
		Env:      env,
		RedisURL: cfgMap.MustGet("redis_url").(string),
		LogLevel: cfgMap.MustGet("log_level").(string),
	}

	return cfg
}
