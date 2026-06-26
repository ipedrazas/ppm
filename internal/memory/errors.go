package memory

import (
	"errors"
	"fmt"
)

// MemoryError is a domain error from the store (bad type, duplicate project,
// missing name on resolve, …), distinct from an underlying I/O error.
type MemoryError struct{ msg string }

func (e *MemoryError) Error() string { return e.msg }

func memErrf(format string, a ...any) error {
	return &MemoryError{msg: fmt.Sprintf(format, a...)}
}

// IsMemoryError reports whether err is a domain MemoryError.
func IsMemoryError(err error) bool {
	var me *MemoryError
	return errors.As(err, &me)
}
