package usecase

import "context"

// UseCase is a command's stateless entrypoint: executed with a context and a
// typed input, it returns a presentation-agnostic *P product or a classified
// *Error. It never renders.
type UseCase[I, P any] interface {
	Execute(ctx context.Context, in I) (*P, error)
}

// Action is a stateless, composable step a use case runs: executed with a
// context and an input I, it returns a *P product or a classified *Error.
type Action[I, P any] interface {
	Execute(ctx context.Context, in I) (*P, error)
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

// NewIOError reports a failure of the underlying storage.
func NewIOError(reason string) *Error {
	return &Error{Type: TypeIO, Reason: reason}
}

// NewNotFoundError reports a named resource that does not exist.
func NewNotFoundError(reason string) *Error {
	return &Error{Type: TypeNotFound, Reason: reason}
}
