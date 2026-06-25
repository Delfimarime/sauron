package usecase

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
)

// DeleteRegistryUseCaseParams injects the collaborators the use case composes.
type DeleteRegistryUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Cascade    *UninstallByRegistryAction
	Logger     *zap.Logger
}

// DeleteRegistryUseCase unregisters a source and cascade-uninstalls its artifacts.
type DeleteRegistryUseCase struct {
	registries storage.RegistriesStore
	cascade    *UninstallByRegistryAction
	logger     *zap.Logger
}

// NewDeleteRegistryUseCase builds the use case from the injected collaborators.
func NewDeleteRegistryUseCase(params DeleteRegistryUseCaseParams) *DeleteRegistryUseCase {
	return &DeleteRegistryUseCase{
		registries: params.Registries,
		cascade:    params.Cascade,
		logger:     params.Logger,
	}
}

// DeleteRegistryInput is the input for deleting a source.
type DeleteRegistryInput struct {
	Name   string
	DryRun bool
}

// DeleteRegistryResult is the outcome of a delete: whether the named registry
// existed, whether the run was a dry-run preview, and the cascade plan.
type DeleteRegistryResult struct {
	Name    string
	Existed bool
	DryRun  bool
	Plan    *DeleteArtifactsByRegistryResponse
}

// Execute runs the find → cascade → dry-run → remove pipeline, returning a *Error
// on the first failing step. A registry that does not exist is a success whose
// result reports nothing existed.
func (uc *DeleteRegistryUseCase) Execute(ctx context.Context, in DeleteRegistryInput) (*DeleteRegistryResult, error) {
	registry, err := uc.registries.FindByName(ctx, in.Name)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("lookup registry %q: %v", in.Name, err))
	}
	if registry == nil {
		uc.logger.Debug("registry not found",
			zap.String(telemetry.FieldRegistryName, in.Name),
		)
		return &DeleteRegistryResult{Name: in.Name, Existed: false}, nil
	}

	plan, err := uc.cascade.Execute(ctx, in.Name)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("uninstall artifacts of %q: %v", in.Name, err))
	}

	if in.DryRun {
		uc.logger.Debug("registry deletion previewed",
			zap.String(telemetry.FieldRegistryName, in.Name),
			zap.Int(telemetry.FieldArtifactCount, plan.Total()),
		)
		return &DeleteRegistryResult{Name: in.Name, Existed: true, DryRun: true, Plan: plan}, nil
	}

	if err := uc.registries.Remove(ctx, in.Name); err != nil {
		return nil, NewIOError(fmt.Sprintf("remove registry %q: %v", in.Name, err))
	}

	uc.logger.Debug("registry removed",
		zap.String(telemetry.FieldRegistryName, in.Name),
		zap.Int(telemetry.FieldArtifactCount, plan.Total()),
	)

	return &DeleteRegistryResult{Name: in.Name, Existed: true, DryRun: false, Plan: plan}, nil
}
