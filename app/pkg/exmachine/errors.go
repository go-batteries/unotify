package exmachine

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidStateIDMapping = errors.New("state_machine_id_mismatch")
	ErrUndeclaredStateID     = errors.New("state_id_alias_mapping_not_found")
)

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
