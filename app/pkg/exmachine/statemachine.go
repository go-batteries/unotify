package exmachine

import (
	"context"
	"errors"
	"unotify/app/pkg/debugtools"
	"unotify/app/pkg/ds"
)

type (
	EventMap      = ds.Map[string, *StateDefinition]
	StateEventMap = ds.Map[string, EventMap]
	ConflictsMap  = ds.Map[string, []*StateDefinition]
)

type StateMachine struct {
	ID string

	reader  StateFileReader
	store   StateEventMap
	aliases AliasMapper
}

type State struct {
	ID    string
	Alias string
}

const (
	WithInvertedAlias = true
	WithAlias         = false
)

var (
	ErrUnaliasedStateName = errors.New("unaliased_state_name")
	ErrStateNotFound      = errors.New("state_not_found_in_machine")
	ErrEventNotRegistered = errors.New("event_not_registered_for_state")
	ErrMalformedState     = errors.New("malformed_event_state_mapping")
)

func (sm *StateMachine) NextState(
	ctx context.Context,
	currState string,
	event string,
	isInverted bool,
) (
	*State,
	bool,
	error,
) {
	if isInverted {
		var ok bool
		currState, ok = sm.aliases.inverseAliasMap[currState]
		if !ok {
			return nil, false, ErrUnaliasedStateName
		}
	}

	evmap, ok := sm.store.Get(currState)
	if !ok {
		return nil, false, ErrStateNotFound
	}

	debugtools.Logdeep(evmap, " finding nemo", event)

	transient, ok := evmap.Get(event)
	if !ok {
		return nil, false, ErrEventNotRegistered
	}

	if transient.Name != currState {
		return nil, false, ErrMalformedState
	}

	if transient.Transition == EndTransitionMarker {
		return nil, true, nil
	}

	alias, ok := sm.aliases.Get(transient.Transition)
	if !ok {
		return nil, false, ErrStateNotFound
	}

	finalState := &State{
		ID:    transient.Transition,
		Alias: alias,
	}

	return finalState, false, nil
}

func Provision(
	ctx context.Context,
	uid string,
	reader StateFileReader,
	filePath string,
) (
	*StateMachine,
	error,
) {
	config, err := reader.Read(ctx, filePath)
	if err != nil {
		return nil, err
	}

	machine := config.Definition
	if machine.Name != uid {
		return nil, ErrInvalidStateIDMapping
	}

	iter := machine.StateIDs.Iter()

	for val, more := iter.Next(); more; val, more = iter.Next() {
		if !config.AliasMap.Has(val) {
			return nil, ErrUndeclaredStateID
		}
	}

	conflicts := ConflictsMap{}
	store := StateEventMap{}
	hasConflicts := false

	stateDefns := machine.States

	for _, state := range stateDefns {
		if !store.Has(state.Name) {
			evmap := ds.NewMap[string, *StateDefinition]()
			event := evmap.Add(state.Event, state)

			store.Add(state.Name, event)
			continue
		}

		// if event exists in store
		evmap, ok := store.Get(state.Name)
		if ok && evmap.Has(state.Event) && handlesConflict(conflicts, state) {
			hasConflicts = true
			continue
		}

		evmap.Add(state.Event, state)
		store.Add(state.Name, evmap)
	}

	if hasConflicts {
		return nil, NewErrFiniteStateViolation(conflicts)
	}

	return &StateMachine{
		ID:      uid,
		store:   store,
		aliases: config.AliasMap,
	}, nil
}

func handlesConflict(conflicts ConflictsMap, defn *StateDefinition) bool {
	if !conflicts.Has(defn.Name) {
		conflicts.Add(defn.Name, []*StateDefinition{defn})
		return true
	}

	conflict, ok := conflicts.Get(defn.Name)
	if ok {
		conflicts.Add(defn.Name, append(conflict, defn))
	}

	return true
}
