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
func (uc *DescribeRegistryUseCase) Execute(ctx context.Context, _ DescribeRegistryInput) (*types.Registry, error) {
	registry, err := uc.registries.Get(ctx)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read registry: %v", err))
	}
	if registry == nil {
		return nil, NewNotFoundError("no registry is configured")
	}

	uc.logger.Debug("registry described",
		zap.String(telemetry.FieldRegistryURI, registry.Spec.Source),
	)

	return registry, nil
}

// DescribeRegistryInput is the per-invocation input for describing the registry.
// Describing the single configured registry takes no business input; field
// selection is a presentation concern resolved by the caller.
type DescribeRegistryInput struct{}
