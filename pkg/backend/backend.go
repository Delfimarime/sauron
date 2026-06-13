// Package backend is the public port for delivery backends.
package backend

import "context"

// Backend is a destination Sauron delivers artifacts to.
type Backend interface {
	// Name returns the backend's stable identifier.
	Name() string
	// Ping verifies the backend is reachable.
	Ping(ctx context.Context) error
}
