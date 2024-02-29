package hookers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/go-playground/validator/v10"
)

const (
	GithubProvider = "github"
)

type Hook struct {
	Provider string `json:"provider" db:"provider" redis:"provider" validate:"required"`
	RepoID   string `json:"repo_id" db:"repo_id" redis:"repo_id" validate:"required"`
	RepoPath string `json:"repo_path" db:"repo_path" redis:"repo_path" validate:"required"`
	Secrets  string `json:"secret" db:"secret" redis:"secret" validate:"required"`
}

func (hook *Hook) Validate() error {
	validate := validator.New()
	return validate.Struct(hook)
}

func (Hook) Table() string {
	return "hookers"
}

func buildProviderKey(hook *Hook) string {
	return fmt.Sprintf("providers::%s", hook.Provider)
}

func buildSecretsKey(hook *Hook) string {
	return fmt.Sprintf("secrets::%s::%s", hook.Provider, hook.RepoID)
}

type GithubHookOpts func(*Hook)

func WithGithubRepoID(repoID string) GithubHookOpts {
	return func(h *Hook) {
		h.RepoID = repoID
	}
}

func WithGithubRepoPath(repoPath string) GithubHookOpts {
	return func(h *Hook) {
		h.RepoPath = repoPath
	}
}

func WithGithubSecretOverride(secret string) GithubHookOpts {
	return func(h *Hook) {
		h.Secrets = secret
	}
}

func NewGithubHook(opts ...GithubHookOpts) (*Hook, error) {
	secret, err := generateRandomHex(32)
	if err != nil {
		return nil, err
	}

	hook := &Hook{
		// ID:       id.String(),
		Provider: GithubProvider,
		Secrets:  secret,
	}

	for _, opt := range opts {
		opt(hook)
	}

	return hook, nil
}

func generateRandomHex(length int) (string, error) {
	randomBytes := make([]byte, length)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomHex := hex.EncodeToString(randomBytes)
	return randomHex, nil
}
