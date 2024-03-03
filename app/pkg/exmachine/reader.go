package exmachine

import (
	"context"
	"errors"
	"unotify/app/pkg/ds"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

const EndTransitionMarker = "<end>"

type StateFileReader interface {
	Read(context.Context, string) (*StateProvisoner, error)
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

	inverseAliasMap map[string]string
}

// Collection of state machines
type StateProvisoner struct {
	Provisioners []*ProvisionerDefinition `hcl:"provisioner,block"`

	provisonerMap map[string]*ProvisionerDefinition
}

func (h *StateProvisoner) Get(key string) (*ProvisionerDefinition, bool) {
	machine, ok := h.provisonerMap[key]
	return machine, ok
}

type ProvisionerDefinition struct {
	Name       string            `hcl:"name,label"`
	Definition MachineDefinition `hcl:"statemachine,block"`
	AliasMap   AliasMapper       `hcl:"aliasmapper,block"`
}

func (a AliasMapper) Has(key string) bool {
	_, ok := a.Aliases[key]
	return ok
}

func (a AliasMapper) Get(key string) (string, bool) {
	value, ok := a.Aliases[key]
	return value, ok
}

func (a AliasMapper) GetInverted(key string) (string, bool) {
	key, _ = a.inverseAliasMap[key]
	value, ok := a.Aliases[key]
	return value, ok
}

var (
	ErrHCLFileParseFailed          = errors.New("file_parse_failed")
	ErrTranistionStateUnregistered = errors.New("transition_state_unregistered")
)

type HCLFileReader struct{}

func (h HCLFileReader) Read(ctx context.Context, filePath string) (
	*StateProvisoner,
	error,
) {
	hype := &StateProvisoner{
		provisonerMap: map[string]*ProvisionerDefinition{},
	}

	err := hclsimple.DecodeFile(filePath, nil, hype)
	if err != nil {
		return nil, err
	}

	// debugtools.Logdeep(hype)

	for _, config := range hype.Provisioners {
		h.SetConfigDefaults(config)

		hype.provisonerMap[config.Name] = config
		definition := config.Definition

		for _, state := range definition.States {
			if state.Transition == EndTransitionMarker {
				continue
			}

			if !definition.StateIDs.Has(state.Transition) {
				return nil, ErrTranistionStateUnregistered
			}
		}
	}

	// debugtools.Logdeep(hype, "hype")
	// debugtools.Logdeep(hype.provisonerMap, "provisonmap")

	return hype, nil
}

func (h HCLFileReader) SetConfigDefaults(config *ProvisionerDefinition) {
	stateIDs := ds.NewSet[string]()
	config.AliasMap.inverseAliasMap = map[string]string{}

	for _, state := range config.Definition.States {
		if state == nil {
			continue
		}

		state.Alias = config.AliasMap.Aliases[state.Name]
		config.AliasMap.inverseAliasMap[state.Alias] = state.Name

		stateIDs.Add(state.Name)
	}

	config.Definition.StateIDs = *stateIDs
}
