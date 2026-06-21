// Package usecase holds Sauron's use cases and the actions they compose.
package usecase

import "go.uber.org/fx"

// NewFxOptions wires the use cases and actions.
func NewFxOptions() fx.Option {
	return fx.Provide(
		NewAddRegistryUseCase,
		NewListRegistriesUseCase,
		NewDescribeRegistryUseCase,
		NewUninstallByRegistryAction,
		NewDeleteRegistryUseCase,
	)
}
