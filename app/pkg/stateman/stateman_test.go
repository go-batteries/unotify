package stateman

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewHclFileReader(t *testing.T) {
	filePath, err := filepath.Abs("../../../config/statemachines/jira.hcl")
	require.NoError(t, err, "should not have failed to find file")

	ctx := context.Background()
	reader := HCLFileReader{}

	config, err := reader.Read(ctx, filePath)
	require.NoError(t, err, "should not have failed to read statemachine config")
	require.NotNil(t, config)

	assert.Equal(t, "soc", config.Definition.Name)
}

func Test_NewStateMachine(t *testing.T) {
	ctx := context.Background()
	filePath, err := filepath.Abs("../../../config/statemachines/jira.hcl")
	require.NoError(t, err, "should not have failed to find file")

	machine, err := BuildMachine(ctx, HCLFileReader{}, "soc", filePath)
	require.NoError(t, err, "should not have failed to initialize state machine")
	require.NotNil(t, machine)

	evmap, err := machine.GetEventMap(ctx, "To Do")
	require.NoError(t, err, "should not have failed to get event map")
	require.NotNil(t, evmap)

	stateData, err := machine.GetStateData(ctx, "To Do")
	require.NoError(t, err, "should not have failed to get state data")

	assert.Equal(t, stateData.Key, "to_do")
	assert.Equal(t, stateData.Alias, "To Do")

	data, err := machine.Transition(ctx, "To Do", "markup")
	require.NoError(t, err)
	require.NotNil(t, data)

	require.Equal(t, data.Alias, "In Progress")

	data, err = machine.Transition(ctx, "In Progress", "markup")
	require.NoError(t, err)
	require.NotNil(t, data)

	require.Equal(t, data.Alias, "Done")

	data, err = machine.Transition(ctx, "In Progress", "markdown")
	require.NoError(t, err)
	require.NotNil(t, data)

	require.Equal(t, data.Alias, "To Do")

	data, err = machine.Transition(ctx, "Done", "markup")
	require.NoError(t, err)
	require.NotNil(t, data)

	require.Equal(t, data.Alias, "Done")
}
