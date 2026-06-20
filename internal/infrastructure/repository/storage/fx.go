package storage

import (
	"go.uber.org/fx"
)

// NewFxOptions wires the storage engine and the typed registries store.
func NewFxOptions() fx.Option {
	return fx.Options(
		fx.Provide(
			newFilesystem,
			NewStore,
			NewRegistriesStore,
		),
	)
}
