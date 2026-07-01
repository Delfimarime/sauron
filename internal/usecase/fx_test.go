package usecase

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// TestNewFxOptions verifies the usecase options compose into a valid container
// once their adapter and collaborator dependencies are supplied.
func TestNewFxOptions(t *testing.T) {
	// Arrange.
	deps := fx.Options(
		fx.Provide(
			fx.Annotate(
				func() extension.Registry { return &extension.MockBasedRegistry{} },
				fx.ResultTags(`name:"registry.git"`),
			),
			fx.Annotate(
				func() extension.Registry { return &extension.MockBasedRegistry{} },
				fx.ResultTags(`name:"registry.http"`),
			),
			fx.Annotate(
				func() afero.Fs { return afero.NewMemMapFs() },
				fx.ResultTags(`name:"provider"`),
			),
			func() storage.RegistriesStore { return &storage.MockBasedRegistriesStore{} },
			func() storage.TrackStore { return &storage.MockBasedTrackStore{} },
			func() storage.ProvidersStore { return &storage.MockBasedProvidersStore{} },
			zap.NewNop,
		),
	)

	// Act.
	app := fx.New(
		deps,
		NewFxOptions(),
		fx.Invoke(func(UseCase[SetRegistryRequest, SetRegistryResponse]) {}),
		fx.Invoke(func(UseCase[DescribeRegistryRequest, types.Registry]) {}),
		fx.Invoke(func(UseCase[ListCatalogueRequest, ListCatalogueResponse]) {}),
		fx.Invoke(func(UseCase[UnsetRegistryRequest, UnsetRegistryResponse]) {}),
		fx.Invoke(func(UseCase[InstallRequest, InstallResponse]) {}),
		fx.Invoke(func(UseCase[DiffRequest, DiffResponse]) {}),
	)

	// Assert.
	assert.NoError(t, app.Err())
}
