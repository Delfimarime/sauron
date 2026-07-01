package usecase

import (
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// NewFxOptions wires the use cases. OpenRegistryUseCase keeps a bespoke interface
// (its Execute returns a filesystem, not a *Response); every other use case is
// provided as its UseCase[I, P] interface, never its concrete type, so a
// consumer — a composing use case (Migrate, Diff) or the command layer's
// runUseCase — depends only on the shape it actually calls. An internal refactor
// of a concrete type (e.g. wrapping the generic ListUseCase, as
// NewListCatalogueUseCase already does) then never ripples out to its callers.
func NewFxOptions() fx.Option {
	return fx.Provide(
		NewOpenRegistryUseCase,
		fx.Annotate(
			NewMigrateUseCase, fx.As(new(UseCase[MigrateRequest, MigrateResponse])),
		),
		fx.Annotate(
			NewDiffUseCase, fx.As(new(UseCase[DiffRequest, DiffResponse])),
		),
		fx.Annotate(
			NewSetRegistryUseCase, fx.As(new(UseCase[SetRegistryRequest, SetRegistryResponse])),
		),
		fx.Annotate(
			NewSetProviderUseCase, fx.As(new(UseCase[SetProviderRequest, SetProviderResponse])),
		),
		fx.Annotate(
			NewDescribeRegistryUseCase, fx.As(new(UseCase[DescribeRegistryRequest, types.Registry])),
		),
		fx.Annotate(
			NewDescribeProviderUseCase, fx.As(new(UseCase[DescribeProviderRequest, types.Provider])),
		),
		fx.Annotate(
			NewListCatalogueUseCase, fx.As(new(UseCase[ListCatalogueRequest, ListCatalogueResponse])),
		),
		fx.Annotate(
			NewInstallUseCase, fx.As(new(UseCase[InstallRequest, InstallResponse])),
		),
		fx.Annotate(
			NewUnsetRegistryUseCase, fx.As(new(UseCase[UnsetRegistryRequest, UnsetRegistryResponse])),
		),
	)
}
