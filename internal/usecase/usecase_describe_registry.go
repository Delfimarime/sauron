package usecase

import (
	"context"

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

// DescribeRegistryUseCase reads the configured registry and returns its detail;
// field selection is a presentation concern resolved by the caller.
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

// Execute runs the get → not-found pipeline, returning a classified *Error on the
// first failing step and otherwise the configured registry.
func (uc *DescribeRegistryUseCase) Execute(ctx context.Context, _ DescribeRegistryRequest) (*types.Registry, error) {
	registry, err := requireRegistry(ctx, uc.registries)
	if err != nil {
		return nil, err
	}

	uc.logger.Debug("registry described",
		zap.String(telemetry.FieldRegistrySource, registry.Spec.Source),
	)

	return registry, nil
}
