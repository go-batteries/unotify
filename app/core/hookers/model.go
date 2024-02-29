package hookers

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/go-playground/validator/v10"
)

const (
	GithubProvider = "github"
)

type Hook struct {
	Provider string `json:"provider" db:"provider" validate:"required"`
	RepoPath string `json:"repo_path" db:"repo_path" validate:"required"`
	Secret   string `json:"secret" db:"secret" validate:"required"`
}

func (hook *Hook) Validate() error {
	validate := validator.New()
	return validate.Struct(hook)
}

func (Hook) Table() string {
	return "hookers"
}

type GithubHookOpts func(*Hook)

func WithGithubRepoPath(repoPath string) GithubHookOpts {
	return func(h *Hook) {
		h.RepoPath = repoPath
	}
}

func WithGithubSecretOverride(secret string) GithubHookOpts {
	return func(h *Hook) {
		h.Secret = secret
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
		Secret:   secret,
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
