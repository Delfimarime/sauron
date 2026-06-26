package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// providerDirs maps a provider name to the home-relative directory its artifacts
// live under.
var providerDirs = map[string]string{
	types.ProviderClaude:   ".claude",
	types.ProviderZencoder: ".zencoder",
}

// MigrateUseCaseParams injects the collaborators the migration composes.
type MigrateUseCaseParams struct {
	fx.In
	Logger *zap.Logger
	Track  storage.TrackStore
	Fs     afero.Fs `name:"provider"`
}

// MigrateUseCase relocates installed artifacts when the provider changes: for
// each recorded artifact it renames the file from the source provider's directory
// to the destination's on the provider filesystem, bumps the recorded updatedAt,
// and updates the track entry. A per-artifact failure is recorded and migration
// continues, leaving the track consistent with what moved. It fits the generic
// UseCase[MigrateInput, MigrateResult] shape that set-provider composes.
type MigrateUseCase struct {
	track  storage.TrackStore
	fs     afero.Fs
	logger *zap.Logger
}

// NewMigrateUseCase builds the use case from the injected collaborators.
func NewMigrateUseCase(params MigrateUseCaseParams) *MigrateUseCase {
	return &MigrateUseCase{
		track:  params.Track,
		fs:     params.Fs,
		logger: params.Logger,
	}
}

// Execute moves every installed artifact from the source to the destination
// provider directory, continuing past per-artifact failures.
func (uc *MigrateUseCase) Execute(ctx context.Context, in MigrateInput) (*MigrateResult, error) {
	artifacts, err := uc.track.List(ctx)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read installed set: %v", err))
	}

	result := &MigrateResult{}
	for _, artifact := range artifacts {
		moved, err := uc.move(ctx, artifact, in)
		if err != nil {
			result.Failures = append(
				result.Failures, MigrateFailure{Artifact: artifact, Reason: err.Error()},
			)
			uc.logger.Warn("artifact migration failed",
				zap.String(telemetry.FieldArtifactName, artifact.Metadata.Name),
				zap.String(telemetry.FieldProviderFrom, in.From),
				zap.String(telemetry.FieldProviderTo, in.To),
				zap.Error(err),
			)
			continue
		}
		result.Moved = append(result.Moved, moved)
	}

	return result, nil
}

// move relocates one artifact's file between provider directories, bumps its
// updatedAt, and persists the track entry. The home-relative spec.path is
// invariant across a provider switch — only the provider home changes.
func (uc *MigrateUseCase) move(ctx context.Context, artifact types.Artifact, in MigrateInput) (types.Artifact, error) {
	src := filepath.Join(providerDirs[in.From], artifact.Spec.Path)
	dst := filepath.Join(providerDirs[in.To], artifact.Spec.Path)

	if err := uc.fs.MkdirAll(filepath.Dir(dst), dirPerm); err != nil {
		return types.Artifact{}, fmt.Errorf("prepare %s: %w", dst, err)
	}
	if err := uc.fs.Rename(src, dst); err != nil {
		return types.Artifact{}, fmt.Errorf("move %s: %w", src, err)
	}

	artifact.Spec.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := uc.track.Update(ctx, artifact); err != nil {
		return types.Artifact{}, fmt.Errorf("update %s: %w", artifact.Metadata.Name, err)
	}

	return artifact, nil
}

// dirPerm is the mode for created provider directories.
const dirPerm os.FileMode = 0o755

// MigrateInput names the source and destination providers by name.
type MigrateInput struct {
	From string
	To   string
}

// MigrateResult is the presentation-agnostic outcome of a migration: the
// artifacts moved (each carrying its Kind) and the per-artifact failures.
type MigrateResult struct {
	Moved    []types.Artifact
	Failures []MigrateFailure
}

// MigrateFailure records one artifact that could not be migrated and why.
type MigrateFailure struct {
	Reason   string
	Artifact types.Artifact
}
