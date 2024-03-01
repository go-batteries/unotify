package stateman

import (
	"context"
	"errors"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type StateFileReader interface {
	Read(context.Context, string) (*MachineConfig, error)
}

// statemachine: { soc: { states: states, transitions: { "to_do": [] } }}
type StateDefinition struct {
	Name  string `hcl:"name,label"`
	Alias string `hcl:"alias"`
	Event string `hcl:"event"`

	Transitions []string `hcl:"transitions"`
}

type MachineDefinition struct {
	Name     string            `hcl:"name,label"`
	StateIDs []string          `hcl:"states"`
	States   []StateDefinition `hcl:"state,block"`
}

type MachineConfig struct {
	Definition MachineDefinition `hcl:"statemachine,block"`
}

var ErrHCLFileParseFailed = errors.New("file_parse_failed")

type HCLFileReader struct{}

func (h HCLFileReader) Read(
	ctx context.Context,
	filePath string,
) (
	*MachineConfig,
	error,
) {
	config := &MachineConfig{}

	err := hclsimple.DecodeFile(filePath, nil, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
