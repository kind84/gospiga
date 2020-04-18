package errors

import (
	"fmt"
)

type ErrDuplicateID struct {
	ID string
}

func (e ErrDuplicateID) Error() string {
	return fmt.Sprintf("ID %s already exists", e.ID)
}
