package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("authentication required")
	ErrForbidden    = errors.New("operation forbidden")
	ErrConflict     = errors.New("resource conflict")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "request validation failed" }

func Invalid(field, message string) error {
	return &ValidationError{Fields: map[string]string{field: message}}
}

type TransitionError struct {
	Entity string
	From   string
	To     string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("invalid %s transition from %q to %q", e.Entity, e.From, e.To)
}

