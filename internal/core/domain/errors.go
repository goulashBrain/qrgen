package domain

import (
	"errors"
	"fmt"
)

// ErrInvalidInput is the root of all client-side errors.
// Adapters map it to e.g. HTTP 400 or a non-zero CLI exit code.
var ErrInvalidInput = errors.New("invalid input")

var (
	ErrEmptyContact    = fmt.Errorf("%w: contact needs a first or last name", ErrInvalidInput)
	ErrNoContacts      = fmt.Errorf("%w: no contacts found (check separator and headers)", ErrInvalidInput)
	ErrTooManyContacts = fmt.Errorf("%w: too many contacts", ErrInvalidInput)
)
