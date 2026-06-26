package usecase

import "go.uber.org/fx"

// NewFxOptions wires the use cases and actions.
func NewFxOptions() fx.Option {
	return fx.Provide(
		fx.Annotate(
			NewOpenRegistryUseCase, fx.As(new(OpenRegistry)),
		),
		NewSetRegistryUseCase,
		NewDescribeRegistryUseCase,
		NewListCatalogueUseCase,
		NewUnsetRegistryUseCase,
	)
}
