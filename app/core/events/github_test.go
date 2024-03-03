package events

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ToGithubEvents(t *testing.T) {
	response := `
	{"retries":0,"event":{"action":"released1","release":{"body":"**Full Changelog**: SOC-1 https://github.com/go-batteries/webhook-test-repo/compare/v1.0.0...v1.0.1","tag_name":"v1.0.1"},"repository":{"full_name":"go-batteries/webhook-test-repo","commits_url":"https://api.github.com/repos/go-batteries/webhook-test-repo/commits{/sha}"}}}`

	ge, err := ParseGithubEvenFromCache([]byte(response))
	require.NoError(t, err, "should have unmarshalled successfully")

	require.Equal(t, ge.Action, "released1")
}
