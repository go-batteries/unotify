package hookers

// func (repo *CacheRepository) Save(ctx context.Context, hook *Hook) error {
// 	if hook == nil {
// 		return ErrEmptyResouce
// 	}
//
// 	if err := hook.Validate(); err != nil {
// 		logrus.WithContext(ctx).WithError(err).Error("failed to validate hook")
// 		return ErrHookValidationFailed
// 	}
//
// 	// TODO: Sanitize the RepoID
//
// 	tx := repo.client.TxPipeline()
// 	if err := tx.SAdd(ctx, buildProviderKey(hook), hook.RepoID).Err(); err != nil {
// 		logrus.WithContext(ctx).WithError(err).Error("failed to add repo to set")
// 		return err
// 	}
//
// 	if err := tx.SAdd(ctx, buildSecretsKey(hook), hook.Secrets).Err(); err != nil {
// 		logrus.WithContext(ctx).WithError(err).Error("failed to add to secrets")
// 		return err
// 	}
//
// 	return nil
// }

// func (repo *CacheRepository) All(ctx context.Context, finder *FindHookByProvider) ([]*Hook, error) {
// 	hook := &Hook{Provider: finder.Provider, RepoID: finder.RepoID}
//
// 	reposForProvider, err := repo.client.SMembers(ctx, buildProviderKey(hook)).Result()
// 	if err != nil {
// 		logrus.WithContext(ctx).WithError(err).Error("failed to repo find repos for ", hook.Provider)
// 		return nil, err
// 	}
//
// 	if len(reposForProvider) == 0 {
// 		return nil, ErrProvidersEmpty
// 	}
//
// 	hooks := []*Hook{}
//
// 	for _, repoPath := range reposForProvider {
// 		hooks = append(hooks, &Hook{Provider: hook.Provider, RepoID: repoPath})
// 	}
//
// 	if !finder.Dive {
// 		return hooks, nil
// 	}
//
// 	// TODO: See if we need some channels
// 	for _, hook := range hooks {
// 		func(h *Hook) {
// 			secrets, err := repo.client.SMembers(ctx, buildSecretsKey(h)).Result()
// 			if err != nil {
// 				logrus.WithContext(ctx).WithError(err).Error("failed to get secrets for key")
// 				return
// 			}
//
// 			h.Secrets = secrets
// 		}(hook)
// 	}
//
// 	return hooks, err
// }
//
// func (repo *CacheRepository) Find(ctx context.Context, finder *FindHookByProvider) (*Hook, error) {
// 	hook := &Hook{Provider: finder.Provider, RepoID: finder.RepoID}
//
// 	hasProviderRegistered, err := repo.client.SIsMember(ctx, buildProviderKey(hook), hook.RepoID).Result()
// 	if err != nil {
// 		logrus.WithContext(ctx).WithError(err).Error("redis ismembers command failed")
// 		return nil, err
// 	}
//
// 	if !hasProviderRegistered {
// 		logrus.WithContext(ctx).Error("webhook for repo not registered")
// 		return nil, ErrHooksRecordEmpty
// 	}
//
// 	secrets, err := repo.client.SMembers(ctx, buildSecretsKey(hook)).Result()
// 	if err != nil {
// 		logrus.WithContext(ctx).WithError(err).Error("failed to get registered secrets")
// 		return nil, err
// 	}
//
// 	hook.Secrets = secrets
// 	if len(hook.Secrets) == 0 {
// 		logrus.WithContext(ctx).Error("no secrets were registerd for github webhook.")
// 		return nil, ErrHooksRecordEmpty
// 	}
//
// 	return hook, err
// }
