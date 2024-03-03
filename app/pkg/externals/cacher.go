package externals

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type ResponseCacher[E any] interface {
	Get(ctx context.Context, key string) (E, bool)
	Set(ctx context.Context, key string, value E, d time.Duration) error
	Delete(ctx context.Context, key string) error
}

type AppCacher[V any] struct {
	client    *cache.Cache
	keyPrefix string
}

var DefaultResponseCacheDuration = 12 * time.Hour

func NewAppCacher[V any](keyPrefix string, expiryDuration time.Duration) *AppCacher[V] {
	c := cache.New(expiryDuration, expiryDuration+(4*time.Minute))

	return &AppCacher[V]{
		client:    c,
		keyPrefix: keyPrefix,
	}
}

func (appcacher *AppCacher[V]) Get(ctx context.Context, key string) (V, bool) {
	var dv V
	key = fmt.Sprintf("%s::%s", appcacher.keyPrefix, key)

	v, ok := appcacher.client.Get(key)
	if !ok {
		return dv, ok
	}

	return v.(V), ok
}

func (appcacher *AppCacher[V]) Set(ctx context.Context, key string, value V, d time.Duration) error {
	key = fmt.Sprintf("%s::%s", appcacher.keyPrefix, key)

	appcacher.client.Set(key, value, d)
	return nil
}

func (appcacher *AppCacher[V]) Delete(ctx context.Context, key string) error {
	key = fmt.Sprintf("%s::%s", appcacher.keyPrefix, key)

	appcacher.client.Delete(key)
	return nil
}
