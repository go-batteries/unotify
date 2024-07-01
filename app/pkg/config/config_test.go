package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ReadWorkerConfig(t *testing.T) {
	filePath := "../../../config/workers.yaml"

	wc, err := GetWorkerConfig(filePath)
	require.NoError(t, err)

	require.Equal(t, 1, len(wc.Workers))

	stateFile := wc.Workers["atlassian"].StateFile
	require.Equal(t, "jira.hcl", stateFile)

	projects := wc.Workers["atlassian"].Projects
	require.ElementsMatch(t, []string{"soc", "devhop"}, projects)
}
