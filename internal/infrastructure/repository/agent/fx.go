package agent

import (
	"github.com/spf13/afero"
	"go.uber.org/fx"

	"github.com/delfimarime/sauron/internal/config"
)

// NewFxOptions wires the provider-scoped filesystem, rooted at the user's home,
// under which the provider artifact directories (.claude, .zencoder) live.
func NewFxOptions() fx.Option {
	return fx.Provide(
		fx.Annotate(
			newProviderFilesystem,
			fx.ResultTags(`name:"provider"`),
		),
	)
}

// newProviderFilesystem builds the provider filesystem rooted at the user's real
// home — distinct from storage's $SAURON_HOME-rooted filesystem.
func newProviderFilesystem(configuration config.Configuration) afero.Fs {
	return afero.NewBasePathFs(afero.NewOsFs(), configuration.UserHomeDirectory)
}
