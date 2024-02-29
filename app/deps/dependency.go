package deps

import (
	"context"
	"riza/app/core/events"
	"riza/app/core/hookers"
	"riza/app/pkg/config"
	"riza/app/resque"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type AppDeps struct {
	HookRegistrationSvc    *hookers.HookerService
	GithubEventsRepository *events.EventsRepository
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
	resqueClient := resque.NewResqueQ(rdb, "events::guthub")

	return &AppDeps{
		HookRegistrationSvc:    hookers.NewHookerService(hookRepo),
		GithubEventsRepository: events.NewGithubEventsRedisRepo(resqueClient),
	}
}
