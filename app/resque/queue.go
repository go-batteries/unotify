package resque

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Queuer interface {
	EnqueueMsg(ctx context.Context, payload Payload) error
	ReadMsg(ctx context.Context, key string) ([]string, error)
}

type Resqueue struct {
	client      *redis.Client
	queuePrefix string
}

type Payload struct {
	Message []byte
	Key     string
}

func NewResqueQ(client *redis.Client, queueName string) *Resqueue {
	return &Resqueue{client: client, queuePrefix: queueName}
}

func (rsq *Resqueue) EnqueueMsg(ctx context.Context, payload Payload) error {
	var err error

	// queueName := fmt.Sprintf("%s::%s", rsq.queuePrefix, payload.Key)
	queueName := rsq.queuePrefix

	err = rsq.client.RPush(ctx, queueName, string(payload.Message)).Err()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to push to queue")
	}

	return err
}

func (rsq *Resqueue) ReadMsg(ctx context.Context, key string) ([]string, error) {
	// queueName := fmt.Sprintf("%s::%s", rsq.queuePrefix, key)
	queueName := rsq.queuePrefix
	return rsq.client.BLPop(ctx, 2*time.Second, queueName).Result()
}
