package registry

import (
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/pkg/sauron/extension"
)

// NewFxOptions wires the registry transport adapters as named extension.Registry
// values, one per transport.
func NewFxOptions() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				newGitFactory,
				fx.As(new(extension.Registry)),
				fx.ResultTags(`name:"registry.git"`),
			),
			fx.Annotate(
				func() extension.Registry { return newRESTFactory() },
				fx.ResultTags(`name:"registry.http"`),
			),
		),
	)
}
