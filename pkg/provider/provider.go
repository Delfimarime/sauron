// Package provider is the public port for AI providers.
package provider

import "context"

// Provider is an AI backend Sauron delegates work to.
type Provider interface {
	// Name returns the provider's stable identifier.
	Name() string
	// Available reports whether the provider can be invoked.
	Available(ctx context.Context) error
}
