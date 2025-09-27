package job

import (
	"errors"
)

// Define domain-specific errors. This isolates the caller from GORM's specific errors.
var (
	// ErrNotFound is returned when a requested record does not exist.
	ErrNotFound = errors.New("job record not found")

	// ErrConflict is returned when an operation violates a unique constraint
	// (e.g., trying to create a job with an ID that already exists).
	ErrConflict = errors.New("job record conflict, probably non-unique ID")

	// ErrInternal is a general error for unexpected failures (e.g., marshaling, DB connection loss).
	ErrInternal = errors.New("internal data access error")
)
