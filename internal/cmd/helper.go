package cmd

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/config"
	"github.com/delfimarime/sauron/internal/infrastructure/backend"
	"github.com/delfimarime/sauron/internal/infrastructure/provider"
	"github.com/delfimarime/sauron/internal/infrastructure/registry"
	"github.com/delfimarime/sauron/internal/infrastructure/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/internal/usecase"
)

// NewApp builds (but does not start) the transversal fx app and appends the caller's opts.
func NewApp(ctx context.Context, opts ...fx.Option) *fx.App {
	base := []fx.Option{
		fx.Supply(ctx),
		telemetry.NewFxOptions(),
		config.NewFxOptions(),
		registry.NewFxOptions(),
		provider.NewFxOptions(),
		backend.NewFxOptions(),
		storage.NewFxOptions(),
		usecase.NewFxOptions(),
		fx.Provide(
			newPondPool,
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	}
	return fx.New(append(base, opts...)...)
}
