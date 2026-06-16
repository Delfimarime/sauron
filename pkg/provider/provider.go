// Package provider is the public port for artifact install destinations.
package provider

import "context"

// Provider is a destination environment where Sauron installs artifacts
// (e.g. claude, zencoder).
type Provider interface {
	// Name returns the provider's stable identifier.
	Name() string
	// Available reports whether the provider destination is usable.
	Available(ctx context.Context) error
}
