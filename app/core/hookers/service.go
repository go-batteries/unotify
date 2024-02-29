package hookers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	Save(ctx context.Context, hook *Hook) error
	All(ctx context.Context, finder *FindHookByProvider) ([]*Hook, error)
	Find(ctx context.Context, finder *FindHookByProvider) (*Hook, error)
}

type CacheRepository struct {
	client *redis.Client
}

func NewHookerCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{client: client}
}

var (
	ErrEmptyResouce         = errors.New("empty_resource")
	ErrHookValidationFailed = errors.New("hook_validation_failed")
	ErrHookNotFound         = errors.New("hook_not_registered")
)

func (repo *CacheRepository) Save(ctx context.Context, hook *Hook) error {
	if hook == nil {
		return ErrEmptyResouce
	}

	if err := hook.Validate(); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to validate hook")
		return ErrHookValidationFailed
	}

	b, err := json.Marshal(hook)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to marshal hook")
		return err
	}

	repo.client.LPush(ctx, hook.Provider, string(b))
	return nil
}

var ErrHooksRecordEmpty = errors.New("empty_records")

func (repo *CacheRepository) All(ctx context.Context, finder *FindHookByProvider) ([]*Hook, error) {
	hooksJSON, err := repo.client.LRange(ctx, finder.Provider, 0, -1).Result()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get registered hook")
		return nil, err
	}

	var hooks []*Hook

	for _, hookJSON := range hooksJSON {
		hook := &Hook{}

		err := json.Unmarshal([]byte(hookJSON), &hook)
		if err != nil {
			return nil, err
		}

		hooks = append(hooks, hook)
	}

	if len(hooks) == 0 {
		err = ErrHooksRecordEmpty
	}

	return hooks, err
}

func (repo *CacheRepository) Find(ctx context.Context, finder *FindHookByProvider) (*Hook, error) {
	hooksJSON, err := repo.client.LRange(ctx, finder.Provider, 0, -1).Result()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get registered hook")
		return nil, err
	}

	for _, hookJSON := range hooksJSON {
		hook := &Hook{}

		err := json.Unmarshal([]byte(hookJSON), &hook)
		if err != nil {
			return nil, err
		}

		if strings.EqualFold(hook.RepoPath, finder.RepoPath) {
			return hook, nil
		}
	}

	return nil, ErrHookNotFound
}

type HookerService struct {
	repo Repository
}

func NewHookerService(repo Repository) *HookerService {
	return &HookerService{repo: repo}
}

var (
	ErrFailedToRegisterHook = errors.New("failed_to_register_hook")
	ErrFailedToPersistHook  = errors.New("failed_to_persist_hook")
)

func (self *HookerService) Register(
	ctx context.Context,
	req *RegisterHookRequest,
) (*RegisterHookerResponse, error) {
	logrus.WithContext(ctx).Infoln("registering web hook")

	hook, err := NewGithubHook(WithGithubRepoPath(req.ProjectPath))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("error failed to generate secret for web hook")
		return nil, ErrFailedToRegisterHook
	}

	if err := self.repo.Save(ctx, hook); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to store to db")
		return nil, ErrFailedToPersistHook
	}

	return &RegisterHookerResponse{
		Secret: hook.Secret,
	}, nil
}

func (self *HookerService) Show(
	ctx context.Context,
	finder *FindHookByProvider,
) (*SearchHookerResponse, error) {
	logrus.WithContext(ctx).Infoln("registering web hook")

	hook, err := self.repo.Find(ctx, finder)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to store to db")
		return nil, ErrFailedToPersistHook
	}

	return &SearchHookerResponse{
		Secret:   hook.Secret,
		RepoPath: hook.RepoPath,
		Provider: hook.Provider,
	}, nil
}

func (self *HookerService) List(
	ctx context.Context,
	finder *FindHookByProvider,
) ([]*SearchHookerResponse, error) {
	logrus.WithContext(ctx).Infoln("registering web hook")

	hooks, err := self.repo.All(ctx, finder)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to store to db")
		return nil, ErrFailedToPersistHook
	}

	resps := []*SearchHookerResponse{}

	for _, hook := range hooks {
		resps = append(resps, &SearchHookerResponse{
			Secret:   hook.Secret,
			RepoPath: hook.RepoPath,
			Provider: hook.Provider,
		})
	}

	return resps, nil
}
