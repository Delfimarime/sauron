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
