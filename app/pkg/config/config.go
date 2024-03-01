package config

import (
	"github.com/go-batteries/diaper"
	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	Env             string
	Port            int
	LogLevel        string
	DatabaseURL     string
	RedisURL        string
	AtlassianAPIKey string
	AtlassianURL    string
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
		Port:            cfgMap.MustGetInt("port"),
		Env:             env,
		RedisURL:        cfgMap.MustGet("redis_url").(string),
		AtlassianAPIKey: cfgMap.MustGet("atlassian_api_key").(string),
		AtlassianURL:    cfgMap.MustGet("atlassian_domain_name").(string),
		LogLevel:        cfgMap.MustGet("log_level").(string),
	}

	return cfg
}
