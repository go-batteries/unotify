package stateman

import (
	"context"
	"encoding/json"
	"fmt"
	"unotify/app/pkg/ds"
)

type (
	EventMap      = ds.Map[string, *StateDefinition]
	StateEventMap = ds.Map[string, EventMap]
	ConflictsMap  = ds.Map[string, []*StateDefinition]
)

type StateMachine struct {
	ID string

	reader StateFileReader
	store  StateEventMap
}

func Provison(
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

	conflicts := ConflictsMap{}
	hasConflicts := false

	defns := config.Definition.States
	store := StateEventMap{}

	for _, defn := range defns {
		event := ds.NewMap[string, *StateDefinition]().Add(defn.Event, defn)

		if !store.Has(defn.Name) {
			store.Add(defn.Name, event)
			continue
		}

		// if event exists in store
		evmap := store.Get(defn.Name)
		if evmap.Has(defn.Event) && handlesConflict(conflicts, defn) {
			hasConflicts = true
		}
	}

	if hasConflicts {
		return nil, NewErrFiniteStateViolation(conflicts)
	}

	return &StateMachine{
		ID:    uid,
		store: store,
	}, nil
}

func handlesConflict(
	conflicts ConflictsMap,
	defn *StateDefinition,
) bool {
	if !conflicts.Has(defn.Name) {
		conflicts.Add(defn.Name, []*StateDefinition{defn})
		return true
	}

	conflicts.Add(defn.Name, append(conflicts.Get(defn.Name), defn))
	return true
}

type ErrFiniteStateViolation struct {
	Err     error
	Message string
}

func (e ErrFiniteStateViolation) Error() string {
	return e.Message
}

func NewErrFiniteStateViolation(conflicts ConflictsMap) ErrFiniteStateViolation {
	var err ErrFiniteStateViolation

	b, merr := json.MarshalIndent(conflicts, "", " ")
	if merr != nil {
		err.Err = err
		err.Message = merr.Error()
		return err
	}

	err.Message = fmt.Sprintf(`"conflicts": %s`, string(b))
	return err
}
