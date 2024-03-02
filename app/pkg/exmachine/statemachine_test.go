package exmachine

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

	expectedAliasMap := map[string]string{
		"to_do":       "To Do",
		"in_progress": "In Progress",
		"done":        "Done",
	}
	assert.Equal(t, "soc", config.Definition.Name)
	assert.Equal(t, "soc", config.AliasMap.Name)
	assert.Equal(t, expectedAliasMap, config.AliasMap.Aliases)
}

type MockFileReader struct {
	config *MachineConfig
}

func (m MockFileReader) Read(ctx context.Context, filePath string) (
	*MachineConfig,
	error,
) {
	(&HCLFileReader{}).SetConfigDefaults(m.config)
	return m.config, nil
}

var aliasMap = map[string]string{
	"to_do":      "To Do",
	"processing": "Processing",
	"done":       "Done",
	"failed":     "Failed",
}

func Test_ProvisionMachine(t *testing.T) {
	ctx := context.Background()

	t.Run(
		"fails if state machine id mismatch",
		func(t *testing.T) {
			invalidConfig := &MachineConfig{
				AliasMap: AliasMapper{Aliases: aliasMap},
				Definition: MachineDefinition{
					Name: "ux",
				},
			}

			sm, err := Provision(
				ctx,
				"devops",
				MockFileReader{invalidConfig},
				"noopPath",
			)

			require.Error(t, err)
			require.Nil(t, sm)
		},
	)

	t.Run(
		"fails if state machine states are not aliased",
		func(t *testing.T) {
			invalidConfig := &MachineConfig{
				AliasMap: AliasMapper{Aliases: aliasMap},
				Definition: MachineDefinition{
					Name: "ux",
					States: []*StateDefinition{
						{
							Name:       "random",
							Event:      "random",
							Transition: "<end>",
						},
					},
				},
			}

			sm, err := Provision(
				ctx,
				"ux",
				MockFileReader{invalidConfig},
				"noopPath",
			)

			require.Error(t, err)
			require.Nil(t, sm)
		},
	)

	t.Run(
		"fails in case of one to multiple state transition",
		func(t *testing.T) {
			invalidMachine := MachineDefinition{
				Name: "invalid",
				States: []*StateDefinition{
					{
						Name:       "to_do",
						Event:      "next",
						Transition: "done",
						Alias:      "To Do",
					},
					{
						Name:       "to_do",
						Event:      "next",
						Transition: "processing",
						Alias:      "To Do",
					},
					{
						Name:       "processing",
						Event:      "next",
						Transition: "Done",
						Alias:      "Processing",
					},
					{
						Name:       "done",
						Event:      "next",
						Transition: "<end>",
						Alias:      "Done",
					},
				},
			}

			inValidConfig := &MachineConfig{
				AliasMap:   AliasMapper{Aliases: aliasMap},
				Definition: invalidMachine,
			}
			sm, err := Provision(
				ctx,
				"invalid",
				MockFileReader{inValidConfig},
				"noopPath",
			)

			require.Error(t, err, "should have provisioned")
			assert.Nil(t, sm)
		})

	t.Run("fails if transition event is not valid", func(t *testing.T) {
		t.Skip()
	})

	t.Run(
		"succeeds for many to one state transitions",
		func(t *testing.T) {
			invalidTransition := MachineDefinition{
				Name: "valid",
				States: []*StateDefinition{
					{
						Name:       "to_do",
						Event:      "next",
						Transition: "done",
						Alias:      "To Do",
					},
					{
						Name:       "to_do",
						Event:      "skiplast",
						Transition: "done",
						Alias:      "To Do",
					},
					{
						Name:       "processing",
						Event:      "next",
						Transition: "Done",
						Alias:      "Processing",
					},
					{
						Name:       "done",
						Event:      "next",
						Transition: "<end>",
						Alias:      "Done",
					},
				},
			}

			inValidConfig := &MachineConfig{
				AliasMap:   AliasMapper{Aliases: aliasMap},
				Definition: invalidTransition,
			}
			sm, err := Provision(
				ctx,
				"valid",
				MockFileReader{inValidConfig},
				"noopPath",
			)

			require.NoError(t, err, "should have provisioned")
			assert.NotNil(t, sm)
		})

	t.Run(
		"success on one to one state transition",
		func(t *testing.T) {
			validTranistion := MachineDefinition{
				Name: "valid",
				States: []*StateDefinition{
					{
						Name:       "to_do",
						Event:      "next",
						Transition: "processing",
						Alias:      "To Do",
					},
					{
						Name:       "processing",
						Event:      "next",
						Transition: "done",
						Alias:      "Processing",
					},
					{
						Name:       "done",
						Event:      "next",
						Transition: "<end>",
						Alias:      "Done",
					},
				},
			}

			validConfig := &MachineConfig{
				AliasMap:   AliasMapper{Aliases: aliasMap},
				Definition: validTranistion,
			}
			sm, err := Provision(
				ctx,
				"valid",
				MockFileReader{validConfig},
				"noopPath",
			)

			require.NoError(t, err, "should have provisioned")
			assert.NotNil(t, sm)
		},
	)
}

func Test_StateTransition(t *testing.T) {
	ctx := context.Background()

	validTranistion := MachineDefinition{
		Name: "valid",
		States: []*StateDefinition{
			{
				Name:       "to_do",
				Event:      "next",
				Transition: "processing",
				Alias:      "To Do",
			},
			{
				Name:       "processing",
				Event:      "next",
				Transition: "done",
				Alias:      "Processing",
			},
			{
				Name:       "processing",
				Event:      "prev",
				Transition: "to_do",
				Alias:      "Processing",
			},
			// {
			// 	Name:       "to_do",
			// 	Event:      "prev",
			// 	Transition: "heaven",
			// 	Alias:      "To Do",
			// },
			{
				Name:       "done",
				Event:      "next",
				Transition: "<end>",
				Alias:      "Done",
			},
		},
	}

	validConfig := &MachineConfig{
		AliasMap:   AliasMapper{Aliases: aliasMap},
		Definition: validTranistion,
	}
	sm, err := Provision(
		ctx,
		"valid",
		MockFileReader{validConfig},
		"noopPath",
	)

	require.NoError(t, err, "should have provisioned")

	t.Run("fails to transition for invalid event", func(t *testing.T) {
		currState := "to_do"

		_, _, err := sm.NextState(ctx, currState, "random", WithAlias)
		require.Error(t, err, "should have failed to get next state")

		assert.IsType(t, ErrEventNotRegistered, err)
	})

	t.Run("returns error if state transitions to invalid state", func(t *testing.T) {
		_, _, err := sm.NextState(ctx, "Do Do", "prev", WithInvertedAlias)
		require.Error(t, err, "should have failed if state transitioned to unregistered state")

		assert.IsType(t, ErrEventNotRegistered, err)
	})

	t.Run("signals end of transition if returned state is terminal", func(t *testing.T) {
		_, ok, err := sm.NextState(ctx, "Done", "next", WithInvertedAlias)
		require.NoError(t, err, "should not have failed to get end state")

		assert.True(t, ok, "should signal end of transition")
	})

	t.Run("returns the prev state object for current state", func(t *testing.T) {
		state, ok, err := sm.
			NextState(ctx, "Processing", "prev", WithInvertedAlias)

		require.NoError(t, err, "should not have failed to fetch next state")

		assert.False(t, ok, "should not have receieved end state")
		assert.Equal(t, state.ID, "to_do")
		assert.Equal(t, state.Alias, "To Do")
	})

	t.Run("returns the next state object for current state", func(t *testing.T) {
		state, ok, err := sm.
			NextState(ctx, "Processing", "next", WithInvertedAlias)

		require.NoError(t, err, "should not have failed to fetch next state")

		assert.False(t, ok, "should not have receieved end state")
		assert.Equal(t, state.ID, "done")
		assert.Equal(t, state.Alias, "Done")
	})
}
