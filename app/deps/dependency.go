package deps

import (
	"context"
	"riza/app/core/hookers"
	"riza/app/pkg/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type AppDeps struct {
	HookRegistrationSvc *hookers.HookerService
}

func BuildAppDeps(cfg *config.AppConfig) *AppDeps {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: "",
		DB:       0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logrus.WithError(err).Fatal("failed to connect to redis")
		return nil
	}

	logrus.Infoln("redis connection success")

	hookRepo := hookers.NewHookerCacheRepository(rdb)

	return &AppDeps{
		HookRegistrationSvc: hookers.NewHookerService(hookRepo),
	}
}
