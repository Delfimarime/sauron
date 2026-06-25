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

// ListRegistriesUseCaseParams injects the collaborators the use case composes.
type ListRegistriesUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Logger     *zap.Logger
}

// ListRegistriesUseCase reads the registered sources.
type ListRegistriesUseCase struct {
	registries storage.RegistriesStore
	logger     *zap.Logger
}

// NewListRegistriesUseCase builds the use case from the injected collaborators.
func NewListRegistriesUseCase(params ListRegistriesUseCaseParams) *ListRegistriesUseCase {
	return &ListRegistriesUseCase{
		registries: params.Registries,
		logger:     params.Logger,
	}
}

// ListRegistriesInput is the parameterless input for the registry listing;
// filtering, sorting, and field selection are view concerns of the client.
type ListRegistriesInput struct{}

// ListRegistriesResult carries the stored registries for the client to render.
type ListRegistriesResult struct {
	Registries []types.Registry
}

// Execute reads the stored registries, returning an io *Error when the state
// cannot be read.
func (uc *ListRegistriesUseCase) Execute(ctx context.Context, _ ListRegistriesInput) (*ListRegistriesResult, error) {
	registries, err := uc.registries.List(ctx)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read registries: %v", err))
	}

	uc.logger.Info("registries listed",
		zap.Int(telemetry.FieldRegistryCount, len(registries)),
	)

	return &ListRegistriesResult{Registries: registries}, nil
}
