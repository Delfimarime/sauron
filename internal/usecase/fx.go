// Package usecase holds Sauron's use cases and the actions they compose.
package usecase

import "go.uber.org/fx"

// NewFxOptions wires the use cases and actions.
func NewFxOptions() fx.Option {
	return fx.Provide(
		fx.Annotate(NewOpenRegistryAction, fx.As(new(OpenRegistry))),
		NewAddRegistryUseCase,
		NewListRegistriesUseCase,
		NewDescribeRegistryUseCase,
		NewListCatalogueUseCase,
		NewUninstallByRegistryAction,
		NewDeleteRegistryUseCase,
	)
}
