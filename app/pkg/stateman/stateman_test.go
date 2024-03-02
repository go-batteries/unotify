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

			sm, err := Provison(
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

			sm, err := Provison(
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
			sm, err := Provison(
				ctx,
				"invalid",
				MockFileReader{inValidConfig},
				"noopPath",
			)

			require.Error(t, err, "should have provisioned")
			assert.Nil(t, sm)
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
			sm, err := Provison(
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

			validConfig := &MachineConfig{
				AliasMap:   AliasMapper{Aliases: aliasMap},
				Definition: validTranistion,
			}
			sm, err := Provison(
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
