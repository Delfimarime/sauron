package cmd

import (
	"context"

	"github.com/alitto/pond/v2"
	"go.uber.org/fx"
)

func newPondPool(ctx context.Context, lc fx.Lifecycle) pond.Pool {
	pool := pond.NewPool(0,
		pond.WithContext(ctx),
	)
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			pool.StopAndWait()
			return nil
		},
	})
	return pool
}
