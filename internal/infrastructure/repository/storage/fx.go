package storage

import (
	"go.uber.org/fx"
)

// NewFxOptions wires the registry adapters.
func NewFxOptions() fx.Option {
	return fx.Options(
		fx.Provide(
			NewStore,
			newFilesystem,
		),
	)
}
