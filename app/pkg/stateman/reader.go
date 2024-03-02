package stateman

import (
	"context"
	"errors"
	"unotify/app/pkg/ds"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type StateFileReader interface {
	Read(context.Context, string) (*MachineConfig, error)
}

type StateDefinition struct {
	Name       string `hcl:"name,label"`
	Event      string `hcl:"event,label"`
	Transition string `hcl:"transition"`

	Alias string
}

type MachineDefinition struct {
	Name       string             `hcl:"name,label"`
	StateIDs   []string           `hcl:"states"`
	States     []*StateDefinition `hcl:"state,block"`
	EntryPoint string             `hcl:"initial"`

	StateIDSet ds.Set[string]
}

type AliasMapper struct {
	Name    string            `hcl:"name,label"`
	Aliases map[string]string `hcl:"aliases"`
}

type MachineConfig struct {
	Definition MachineDefinition `hcl:"statemachine,block"`
	AliasMap   AliasMapper       `hcl:"aliasmapper,block"`
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

	for _, state := range config.Definition.States {
		if state == nil {
			continue
		}

		state.Alias = config.AliasMap.Aliases[state.Name]
	}

	config.Definition.StateIDSet = *ds.ToSet[string](config.Definition.StateIDs...)

	return config, nil
}
