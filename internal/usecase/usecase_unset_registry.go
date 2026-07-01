package usecase

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
)

// UnsetRegistryUseCaseParams injects the collaborators the use case composes.
type UnsetRegistryUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Logger     *zap.Logger
}

// UnsetRegistryUseCase removes the configured registry. Installed artifacts are
// preserved — unsetting the registry never uninstalls anything.
type UnsetRegistryUseCase struct {
	registries storage.RegistriesStore
	logger     *zap.Logger
}

// NewUnsetRegistryUseCase builds the use case from the injected collaborators.
func NewUnsetRegistryUseCase(params UnsetRegistryUseCaseParams) *UnsetRegistryUseCase {
	return &UnsetRegistryUseCase{
		registries: params.Registries,
		logger:     params.Logger,
	}
}

// Execute runs the get → dry-run → remove pipeline, returning a *Error on the
// first failing step. Unsetting when no registry is configured reports that
// nothing was unset and succeeds.
func (uc *UnsetRegistryUseCase) Execute(ctx context.Context, in UnsetRegistryRequest) (*UnsetRegistryResponse, error) {
	registry, err := uc.registries.Get(ctx)
	if err != nil {
		return nil, ioErr("read registry", err)
	}
	if registry == nil {
		uc.logger.Debug("no registry configured")
		return &UnsetRegistryResponse{Outcome: UnsetNothing}, nil
	}

	if in.DryRun {
		uc.logger.Debug("registry unset previewed")
		return &UnsetRegistryResponse{Outcome: UnsetPreview}, nil
	}

	if err := uc.registries.Remove(ctx); err != nil {
		return nil, ioErr("remove registry", err)
	}

	uc.logger.Info("registry unset", zap.String(telemetry.FieldRegistrySource, registry.Spec.Source))

	return &UnsetRegistryResponse{Outcome: UnsetRemoved}, nil
}
