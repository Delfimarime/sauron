package usecase

import (
	"context"
	"fmt"

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
func (uc *UnsetRegistryUseCase) Execute(ctx context.Context, in UnsetRegistryInput) (*UnsetRegistryResult, error) {
	registry, err := uc.registries.Get(ctx)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read registry: %v", err))
	}
	if registry == nil {
		uc.logger.Debug("no registry configured")
		return &UnsetRegistryResult{Outcome: UnsetNothing}, nil
	}

	if in.DryRun {
		uc.logger.Debug("registry unset previewed")
		return &UnsetRegistryResult{Outcome: UnsetPreview}, nil
	}

	if err := uc.registries.Remove(ctx); err != nil {
		return nil, NewIOError(fmt.Sprintf("remove registry: %v", err))
	}

	uc.logger.Info("registry unset", zap.String(telemetry.FieldRegistryURI, registry.Spec.Source))

	return &UnsetRegistryResult{Outcome: UnsetRemoved}, nil
}

// UnsetOutcome classifies which removal outcome occurred, so the client can
// render the matching report.
type UnsetOutcome string

// The outcomes unsetting the registry can produce.
const (
	// UnsetNothing reports no registry was configured, so nothing was unset.
	UnsetNothing UnsetOutcome = "nothing"
	// UnsetPreview reports a dry-run preview that changed no state.
	UnsetPreview UnsetOutcome = "preview"
	// UnsetRemoved reports the configured registry was removed.
	UnsetRemoved UnsetOutcome = "removed"
)

// UnsetRegistryResult is the presentation-agnostic outcome of unsetting.
type UnsetRegistryResult struct {
	Outcome UnsetOutcome
}

// UnsetRegistryInput is the per-invocation input for removing the registry.
type UnsetRegistryInput struct {
	DryRun bool
}
