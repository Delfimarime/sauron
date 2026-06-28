package usecase

import "go.uber.org/fx"

// NewFxOptions wires the use cases. OpenRegistryUseCase keeps a bespoke interface
// (its Execute returns a filesystem, not a *Result); the migrate step fits the
// generic UseCase shape and is provided as UseCase[MigrateInput, MigrateResult].
func NewFxOptions() fx.Option {
	return fx.Provide(
		NewOpenRegistryUseCase,
		fx.Annotate(
			NewMigrateUseCase, fx.As(new(UseCase[MigrateInput, MigrateResult])),
		),
		fx.Annotate(
			NewDiffUseCase, fx.As(new(UseCase[DiffInput, Diff])),
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
