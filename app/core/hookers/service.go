package hookers

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Repository interface {
	Save(ctx context.Context, hook *Hook) error
	Update(ctx context.Context, hook *Hook) error
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

// map[string]map[string]string

// SADD providers::github repo_a repo_b
// SADD secrets::github::repo_a secret_1 secret_2
// Nah, this is a problem, multiple secrets per repo not allowed.
// Switch
// HSET providers::github repo_a secret_a repo_b secret_b
// SADD providers::github repo_a repo_b
var (
	ErrProvidersEmpty   = errors.New("empty_providers")
	ErrHooksRecordEmpty = errors.New("empty_records")
	ErrDuplicateRecord  = errors.New("duplicate_record_creation")
)

func (repo *CacheRepository) Save(ctx context.Context, hook *Hook) error {
	if hook == nil {
		return ErrEmptyResouce
	}

	if err := hook.Validate(); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to validate hook")
		return ErrHookValidationFailed
	}

	if ok, _ := repo.client.SIsMember(ctx, buildProviderKey(hook), hook.RepoID).Result(); ok {
		logrus.WithContext(ctx).Infoln("a hook is already registered. You need to be superadmin, to change code")
		return ErrDuplicateRecord
	}

	if err := repo.client.SAdd(ctx, buildProviderKey(hook), hook.RepoID).Err(); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to add repo path to provider")
		return err
	}

	return repo.client.HSet(ctx, buildSecretsKey(hook), hook).Err()
}

func (repo *CacheRepository) Update(ctx context.Context, hook *Hook) error {
	if hook == nil {
		return ErrEmptyResouce
	}

	if err := hook.Validate(); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to validate hook")
		return ErrHookValidationFailed
	}

	if ok, _ := repo.client.SIsMember(ctx, buildProviderKey(hook), hook.RepoID).Result(); ok {
		logrus.WithContext(ctx).Infoln("a hook is already registered. Overrriding")
	}

	if err := repo.client.SAdd(ctx, buildProviderKey(hook), hook.RepoID).Err(); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to add repo path to provider")
		return err
	}

	return repo.client.HSet(ctx, buildSecretsKey(hook), hook).Err()
}

func (repo *CacheRepository) All(ctx context.Context, finder *FindHookByProvider) ([]*Hook, error) {
	hook := &Hook{
		Provider: finder.Provider,
	}

	webhookRepos, err := repo.client.SMembers(ctx, buildProviderKey(hook)).Result()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get webhooks for provider ", hook.Provider)
		return nil, err
	}

	hooks := []*Hook{}

	for _, repo := range webhookRepos {
		h := &Hook{
			Provider: finder.Provider,
			RepoID:   repo,
		}

		hooks = append(hooks, h)
	}

	if len(hooks) == 0 {
		return nil, ErrHooksRecordEmpty
	}

	if !finder.Dive {
		return hooks, nil
	}

	for _, hook := range hooks {
		func(h *Hook) {
			err := repo.client.HGetAll(ctx, buildSecretsKey(h)).Scan(h)
			if err != nil {
				logrus.WithContext(ctx).WithError(err).Error("secrets not configured for the webhook", buildSecretsKey(h))
				return
			}
		}(hook)
	}

	logrus.WithContext(ctx).Debugf("list repos %+v\n", hooks)

	return hooks, nil
}

func (repo *CacheRepository) Find(ctx context.Context, finder *FindHookByProvider) (*Hook, error) {
	hook := &Hook{
		Provider: finder.Provider,
		RepoID:   finder.RepoID,
	}

	isWebHookRegistered, err := repo.client.SIsMember(ctx, buildProviderKey(hook), hook.RepoID).Result()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("redis call failed to find hook")
		return nil, err
	}

	if !isWebHookRegistered {
		return nil, ErrHookNotFound
	}

	err = repo.client.HGetAll(ctx, buildSecretsKey(hook)).Scan(hook)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get secrets for key")
	}

	logrus.WithContext(ctx).Debugf("repo details %+v", hook)

	return hook, err
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

func (svc *HookerService) Register(
	ctx context.Context,
	req *RegisterHookRequest,
) (*RegisterHookerResponse, error) {
	logrus.WithContext(ctx).Infoln("registering web hook")

	hook, err := NewGithubHook(
		WithGithubRepoPath(req.RepoPath),
		WithGithubRepoID(req.RepoID),
	)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("error failed to generate secret for web hook")
		return nil, ErrFailedToRegisterHook
	}

	if req.ForceUpdate {
		err = svc.repo.Update(ctx, hook)
	} else {
		err = svc.repo.Save(ctx, hook)
	}
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to store to db")
		return nil, ErrFailedToPersistHook
	}

	return &RegisterHookerResponse{
		Secret: hook.Secrets,
	}, nil
}

func (svc *HookerService) FindByRepoProvider(
	ctx context.Context,
	finder *FindHookByProvider,
) (*SearchHookerResponse, error) {
	logrus.WithContext(ctx).Infoln("registering web hook")

	hook, err := svc.repo.Find(ctx, finder)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to store to db")
		return nil, ErrFailedToPersistHook
	}

	return &SearchHookerResponse{
		Secrets:  hook.Secrets,
		RepoID:   hook.RepoID,
		Provider: hook.Provider,
	}, nil
}

func (svc *HookerService) List(
	ctx context.Context,
	finder *FindHookByProvider,
) ([]*SearchHookerResponse, error) {
	logrus.WithContext(ctx).Infoln("registering web hook")

	hooks, err := svc.repo.All(ctx, finder)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to store to db")
		return nil, ErrFailedToPersistHook
	}

	resps := []*SearchHookerResponse{}

	for _, hook := range hooks {
		resps = append(resps, &SearchHookerResponse{
			Secrets:  hook.Secrets,
			RepoID:   hook.RepoID,
			Provider: hook.Provider,
		})
	}

	return resps, nil
}
