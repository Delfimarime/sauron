package storage

import (
	"github.com/delfimarime/sauron/internal/config"
	"github.com/spf13/afero"
)

func newFilesystem(configuration config.Configuration) afero.Fs {
	return afero.NewBasePathFs(
		afero.NewOsFs(), configuration.HomeDirectory,
	)
}
