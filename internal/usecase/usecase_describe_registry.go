package usecase

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DescribeRegistryUseCaseParams injects the collaborators the use case composes.
type DescribeRegistryUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Logger     *zap.Logger
}

// DescribeRegistryUseCase reads one registry by name.
type DescribeRegistryUseCase struct {
	registries storage.RegistriesStore
	logger     *zap.Logger
}

// NewDescribeRegistryUseCase builds the use case from the injected collaborators.
func NewDescribeRegistryUseCase(params DescribeRegistryUseCaseParams) *DescribeRegistryUseCase {
	return &DescribeRegistryUseCase{
		registries: params.Registries,
		logger:     params.Logger,
	}
}

// DescribeRegistryInput is the input for describing one registry.
type DescribeRegistryInput struct {
	Name string
}

// Execute reads the named registry, returning an io *Error when the state cannot
// be read and a not-found *Error when no registry carries the name.
func (uc *DescribeRegistryUseCase) Execute(ctx context.Context, in DescribeRegistryInput) (*types.Registry, error) {
	registry, err := uc.registries.FindByName(ctx, in.Name)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read registry %q: %v", in.Name, err))
	}
	if registry == nil {
		return nil, NewNotFoundError(fmt.Sprintf("registry %q does not exist", in.Name))
	}

	uc.logger.Debug("registry described",
		zap.String(telemetry.FieldRegistryName, registry.Metadata.Name),
	)

	return registry, nil
}
