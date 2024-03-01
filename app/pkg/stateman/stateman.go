package stateman

import (
	"context"
	"errors"
	"unotify/app/pkg/ds"

	"github.com/sirupsen/logrus"
)

// StateMachine
//
//	ID: unique id, in our case jira project name
//	Reader: config reader to read state definitions
//	store:  holds internal state data to answer queries
type StateMachine struct {
	ID     string
	Reader StateFileReader

	store StateStore
}

type StateData struct {
	Key         string // in hcl -> state "key"
	Alias       string
	Event       string
	Transitions *ds.Set[string]
	Terminal    bool
}

// "markup": {}
type StateEventMap = ds.Map[string, StateData]

// "state_id": { "markup": {}, "markdown": {} }
type States map[string]StateEventMap

type StateStore struct {
	States       States
	inverseMap   map[string]string
	stateDataMap map[string]StateData
}

func (s StateStore) NextState(
	prevState,
	event string,
) (
	*StateData,
	error,
) {
	evmap, ok := s.States[prevState]
	if !ok {
		logrus.Error("state is not in config")
		return nil, errors.New("unregistered_state")
	}

	state := evmap.Get(event)
	if len(state.Key) == 0 {
		logrus.Error(event, " event is unregistered")
		return nil, errors.New("unregistered_event")
	}

	logrus.Printf("%s %s transitions %+v", state.Alias, state.Key, state.Transitions)

	nextStateID := state.Transitions.Get(0)
	if len(nextStateID) == 0 {
		logrus.Error("failed to get state id from transitions")
		return &StateData{}, nil
	}

	evmap, ok = s.States[nextStateID]
	if len(nextStateID) == 0 {
		logrus.Error(nextStateID, " eventMap not found in state store")
		return &StateData{}, nil
	}

	nextState, _ := s.stateDataMap[nextStateID]
	return &nextState, nil
}

var ErrMismatchMachineConfig = errors.New("invalid_machine_id")

// Trying this probably horrible coding style
// Because 80 line per character width is apparently
// a magic number, everyone thinks is cool

// Build, the state machine from config.
//
//	each project gets its own state
//	machine, id-ed by ID

func BuildMachine(
	ctx context.Context,
	reader StateFileReader,
	id string,
	configFilePath string,
) (
	*StateMachine,
	error,
) {
	sm := &StateMachine{
		ID:     id,
		Reader: reader,
	}

	config, err := sm.
		Reader.
		Read(ctx, configFilePath)
	if err != nil {
		logrus.
			WithContext(ctx).
			WithError(err).
			Error("failed to read confif file")
	}

	logrus.
		WithContext(ctx).
		Debugf("state machine loaded %v\n", config)

	machineDefn := config.Definition

	if machineDefn.Name != sm.ID {
		logrus.
			WithContext(ctx).
			Error("state machine id and config definition name mismatch")

		return nil, ErrMismatchMachineConfig
	}

	eventmaps := map[string]StateEventMap{}
	inverseMap := map[string]string{}
	stateMap := map[string]StateData{}

	for _, state := range machineDefn.States {
		data := StateData{
			Key:         state.Name,
			Alias:       state.Alias,
			Event:       state.Event,
			Transitions: ds.ToSet(state.Transitions...),
		}

		data.Terminal = data.Transitions.Get(0) == "-"

		if _, ok := eventmaps[data.Key]; !ok {
			eventmaps[data.Key] = ds.Map[string, StateData]{}
		}

		eventmaps[data.Key].Add(state.Event, data)
		stateMap[data.Key] = data

		inverseMap[data.Alias] = data.Key

		logrus.Printf("transitions %v+\n", data.Transitions)
	}

	sm.store.States = eventmaps
	sm.store.inverseMap = inverseMap
	sm.store.stateDataMap = stateMap

	return sm, nil
}

func (sm *StateMachine) GetEventMap(
	ctx context.Context,
	stateAlias string,
) (StateEventMap, error) {
	stateID, ok := sm.store.inverseMap[stateAlias]
	if !ok {
		return nil, errors.New("unregistered_event")
	}

	evmap, ok := sm.store.States[stateID]
	if !ok {
		return nil, errors.New("state_not_registered")
	}

	return evmap, nil
}

func (sm *StateMachine) GetStateData(
	ctx context.Context,
	stateAlias string,
) (
	*StateData,
	error,
) {
	stateID, ok := sm.store.inverseMap[stateAlias]
	if !ok {
		return nil, errors.New("unregistered_event")
	}

	value, ok := sm.store.stateDataMap[stateID]
	if !ok {
		return nil, errors.New("state_not_registered")
	}

	return &value, nil
}

func (sm *StateMachine) Transition(
	ctx context.Context,
	fromStateAlias string,
	event string,
) (
	*StateData,
	error,
) {
	nowStateID, ok := sm.store.inverseMap[fromStateAlias]
	if !ok {
		logrus.
			WithContext(ctx).
			Error("event alias ", fromStateAlias, "not found in state")

		return nil, errors.New("unregistered_event_alias")
	}

	return sm.store.NextState(nowStateID, event)
}
