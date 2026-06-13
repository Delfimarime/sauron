package storage

import (
	"go.uber.org/fx"
)

// NewFxOptions wires the home-rooted filesystem and the stores over it.
func NewFxOptions() fx.Option {
	return fx.Options(
		fx.Provide(
			NewStore,
			newFilesystem,
		),
	)
}
