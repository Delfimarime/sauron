package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
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
			func() storage.RegistriesStore { return &storage.MockBasedRegistriesStore{} },
			zap.NewNop,
		),
	)

	// Act.
	app := fx.New(
		deps,
		NewFxOptions(),
		fx.Invoke(func(*SetRegistryUseCase) {}),
		fx.Invoke(func(*DescribeRegistryUseCase) {}),
		fx.Invoke(func(*ListCatalogueUseCase) {}),
		fx.Invoke(func(*UnsetRegistryUseCase) {}),
	)

	// Assert.
	assert.NoError(t, app.Err())
}
