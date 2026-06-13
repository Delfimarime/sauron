// Package registry is the public port for artifact sources.
package registry

import "context"

// Registry is an artifact source Sauron can read from.
type Registry interface {
	// Name returns the registry's stable identifier.
	Name() string
	// Ping verifies the source is reachable.
	Ping(ctx context.Context) error
}
