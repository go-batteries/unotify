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
	States     []*StateDefinition `hcl:"state,block"`
	EntryPoint string             `hcl:"initial"`

	StateIDs ds.Set[string]
}

type AliasMapper struct {
	Name    string            `hcl:"name,label"`
	Aliases map[string]string `hcl:"aliases"`
}

func (a AliasMapper) Has(key string) bool {
	_, ok := a.Aliases[key]
	return ok
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

	h.SetConfigDefaults(config)

	return config, nil
}

func (h HCLFileReader) SetConfigDefaults(config *MachineConfig) {
	stateIDs := ds.NewSet[string]()
	for _, state := range config.Definition.States {
		if state == nil {
			continue
		}

		state.Alias = config.AliasMap.Aliases[state.Name]
		stateIDs.Add(state.Name)
	}

	config.Definition.StateIDs = *stateIDs
}
