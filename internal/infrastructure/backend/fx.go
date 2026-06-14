// Package backend wires the backend adapters implementing the pkg/backend port.
package backend

import "go.uber.org/fx"

// NewFxOptions wires the backend adapters.
func NewFxOptions() fx.Option {
	return fx.Options()
}
