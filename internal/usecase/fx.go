package usecase

import "go.uber.org/fx"

// NewFxOptions wires the use cases. OpenRegistryUseCase keeps a bespoke interface
// (its Execute returns a filesystem, not a *Response); the migrate step fits the
// generic UseCase shape and is provided as UseCase[MigrateRequest, MigrateResponse].
func NewFxOptions() fx.Option {
	return fx.Provide(
		NewOpenRegistryUseCase,
		fx.Annotate(
			NewMigrateUseCase, fx.As(new(UseCase[MigrateRequest, MigrateResponse])),
		),
		fx.Annotate(
			NewDiffUseCase, fx.As(new(UseCase[DiffRequest, DiffResponse])),
		),
		NewSetRegistryUseCase,
		NewSetProviderUseCase,
		NewDescribeRegistryUseCase,
		NewDescribeProviderUseCase,
		NewListCatalogueUseCase,
		NewInstallUseCase,
		NewUnsetRegistryUseCase,
	)
}
