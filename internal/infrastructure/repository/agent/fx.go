// Package agent wires the provider adapters implementing the extension.Provider
// port (pkg/sauron/extension).
package agent

import "go.uber.org/fx"

// NewFxOptions wires the provider adapters.
func NewFxOptions() fx.Option {
	return fx.Options()
}
