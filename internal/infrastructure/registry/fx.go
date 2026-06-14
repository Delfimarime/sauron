// Package registry wires the registry adapters implementing the pkg/registry port.
package registry

import "go.uber.org/fx"

// NewFxOptions wires the registry adapters.
func NewFxOptions() fx.Option {
	return fx.Options()
}
