package usecase

import (
	"context"
	"io"
)

// Request is the per-invocation context object: a context.Context that also exposes the command's output writer.
type Request interface {
	context.Context
	// Out returns the writer the command's output goes to.
	Out() io.Writer
}

// UseCase is a command's stateless entrypoint, executed with a Request.
type UseCase[R Request] interface {
	Execute(request R) error
}

// Action is a reusable step a use case composes, taking an explicit context and a plain input.
type Action[R, P any] interface {
	Execute(ctx context.Context, input R) (*P, error)
}

// Type classifies a use-case failure so callers can map it to an exit code or
// presentation without inspecting the reason text.
type Type string

// The classes a use case attaches to a failure.
const (
	// TypeUsage marks a caller mistake that can be corrected by changing input.
	TypeUsage Type = "usage"
	// TypeConflict marks a request that collides with existing state.
	TypeConflict Type = "conflict"
	// TypeUnreachable marks a source that could not be reached or read.
	TypeUnreachable Type = "unreachable"
	// TypeValidation marks input that is well-formed but semantically invalid.
	TypeValidation Type = "validation"
	// TypeIO marks a failure of the underlying storage.
	TypeIO Type = "io"
	// TypeNotFound marks a named resource that does not exist.
	TypeNotFound Type = "not_found"
)

// Error is a classified use-case failure: Type buckets it, Reason explains it.
type Error struct {
	Type   Type
	Reason string
}

// Error returns the human-readable reason.
func (e *Error) Error() string {
	return e.Reason
}

// NewUsageError reports a correctable caller mistake.
func NewUsageError(reason string) *Error {
	return &Error{Type: TypeUsage, Reason: reason}
}

// NewConflictError reports a collision with existing state.
func NewConflictError(reason string) *Error {
	return &Error{Type: TypeConflict, Reason: reason}
}

// NewUnreachableError reports a source that could not be reached or read.
func NewUnreachableError(reason string) *Error {
	return &Error{Type: TypeUnreachable, Reason: reason}
}

// NewValidationError reports input that is well-formed but semantically invalid.
func NewValidationError(reason string) *Error {
	return &Error{Type: TypeValidation, Reason: reason}
}

// NewIOError reports a failure of the underlying storage.
func NewIOError(reason string) *Error {
	return &Error{Type: TypeIO, Reason: reason}
}

// NewNotFoundError reports a named resource that does not exist.
func NewNotFoundError(reason string) *Error {
	return &Error{Type: TypeNotFound, Reason: reason}
}
