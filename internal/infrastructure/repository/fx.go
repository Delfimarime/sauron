package repository

import (
	"github.com/delfimarime/sauron/internal/infrastructure/repository/agent"
	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry"
	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"go.uber.org/fx"
)

// NewFxOptions wires the home-rooted filesystem and the stores over it.
func NewFxOptions() fx.Option {
	return fx.Options(
		storage.NewFxOptions(),
		registry.NewFxOptions(),
		agent.NewFxOptions(),
	)
}
