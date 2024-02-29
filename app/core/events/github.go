package events

import (
	"encoding/json"
	"riza/app/resque"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type GithubEvents struct {
	Action     string              `json:"action"`
	Release    GithubEventsRelease `json:"release"`
	Repository GithubRepository    `json:"repository"`
}

type GithubEventsRelease struct {
	Body string `json:"body"`
	Tag  string `json:"tag_name"`
}

type GithubRepository struct {
	FullName   string `json:"full_name"`
	CommitsURL string `json:"commits_url"`
}

func ToGithubEvents(b []byte) (*GithubEvents, error) {
	ev := &GithubEvents{}

	if err := json.Unmarshal(b, ev); err != nil {
		return nil, err
	}

	return ev, nil
}

type EventsRepository struct {
	rsqcl resque.Queuer
}

func NewGithubEventsRedisRepo(rsqcl resque.Queuer) *EventsRepository {
	return &EventsRepository{
		rsqcl: rsqcl,
	}
}

// events store.
// we will use rpush
// and then wait for pop with BLPOP
type EventMessage struct {
	Retries int           `json:"retries"`
	Event   *GithubEvents `json:"event"`
}

func (er *EventsRepository) Create(ctx context.Context, key string, event *GithubEvents) error {
	msg := EventMessage{
		Retries: 0,
		Event:   event,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to marshal event msg")
		return err
	}

	err = er.rsqcl.EnqueueMsg(ctx, resque.Payload{Message: b, Key: key})
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to push event msg to queue")
	} else {
		logrus.WithContext(ctx).Infoln("successfully pushed to queue", key)
	}

	return nil
}
