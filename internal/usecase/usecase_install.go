package usecase

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// InstallUseCaseParams injects the stores and collaborators the install composes.
type InstallUseCaseParams struct {
	fx.In
	Logger     *zap.Logger
	Providers  storage.ProvidersStore
	Registries storage.RegistriesStore
	Track      storage.TrackStore
	Open       OpenRegistryUseCase
	Diff       UseCase[DiffRequest, DiffResponse]
	Fs         afero.Fs `name:"provider"`
}

// InstallUseCase installs named artifacts of one kind from the configured
// registry into the active provider: it requires a provider (FR-005), opens the
// registry (FR-007), resolves each name's desired version cheaply from the source
// listing (no content fetch), composes the diff to plan additions and updates —
// the diff carries the prior tracked artifact for each update so installedAt is
// preserved without a second track read — then fetches and writes each planned
// artifact and records its track document. A name the registry does not offer is
// reported and the run continues (FR-006). It fits the generic
// UseCase[InstallRequest, InstallResponse] shape and never renders.
type InstallUseCase struct {
	fs         afero.Fs
	logger     *zap.Logger
	providers  storage.ProvidersStore
	registries storage.RegistriesStore
	track      storage.TrackStore
	open       OpenRegistryUseCase
	diff       UseCase[DiffRequest, DiffResponse]
}

// NewInstallUseCase builds the use case from the injected stores and collaborators.
func NewInstallUseCase(params InstallUseCaseParams) *InstallUseCase {
	return &InstallUseCase{
		logger:     params.Logger,
		providers:  params.Providers,
		registries: params.Registries,
		track:      params.Track,
		open:       params.Open,
		diff:       params.Diff,
		fs:         params.Fs,
	}
}

// Execute requires a provider, opens the registry, resolves desired versions,
// diffs them against the recorded set, and applies the resulting add/update plan,
// continuing past per-name failures.
func (uc *InstallUseCase) Execute(ctx context.Context, in InstallRequest) (*InstallResponse, error) {
	if err := uc.validate(in); err != nil {
		return nil, err
	}

	home, err := uc.providerHome(ctx)
	if err != nil {
		return nil, err
	}

	fs, err := uc.openRegistry(ctx)
	if err != nil {
		return nil, err
	}

	result := &InstallResponse{}
	desired, err := uc.resolve(ctx, in, fs, result)
	if err != nil {
		return nil, err
	}

	plan, err := uc.diff.Execute(ctx, DiffRequest{Desired: desired})
	if err != nil {
		return nil, err
	}

	uc.apply(ctx, installRun{kind: in.Kind, home: home, fs: fs}, plan, result)
	uc.logger.Info("install reconciled",
		zap.Int(telemetry.FieldArtifactCount, len(result.Added)+len(result.Updated)),
	)

	return result, nil
}

// validate rejects a kind outside the installable roots before any store read.
func (uc *InstallUseCase) validate(in InstallRequest) error {
	if _, ok := artifactKindDirs[in.Kind]; !ok {
		return NewUsageError(fmt.Sprintf("unknown kind %q", in.Kind))
	}

	return nil
}

// providerHome resolves the active provider's home-relative directory, failing
// with a non-usage runtime error when no provider is set (FR-005).
func (uc *InstallUseCase) providerHome(ctx context.Context) (string, error) {
	provider, err := uc.providers.Get(ctx)
	if err != nil {
		return "", ioErr("read provider", err)
	}
	if provider == nil {
		return "", NewNotFoundError("no provider is set; set one before installing")
	}

	return providerDirs[provider.Metadata.Name], nil
}

// openRegistry resolves the single registry and opens its source live; an
// unreachable source surfaces as a runtime error (FR-007).
func (uc *InstallUseCase) openRegistry(ctx context.Context) (source.FileSystem, error) {
	registry, err := requireRegistry(ctx, uc.registries)
	if err != nil {
		return nil, err
	}

	return uc.open.Execute(ctx, *registry)
}

// resolve lists every entry the kind offers and maps each requested name to its
// desired version, recording a per-name failure for a name the listing does not
// offer (FR-006) or whose version the source declares empty (versioning FR-005).
func (uc *InstallUseCase) resolve(
	ctx context.Context, in InstallRequest, fs source.FileSystem, result *InstallResponse,
) ([]DesiredArtifact, error) {
	entries, err := listAll(ctx, fs, artifactKindDirs[in.Kind])
	if err != nil {
		return nil, classifyAdapterErr(err)
	}

	offered := make(map[string]string, len(entries))
	for _, entry := range entries {
		offered[catalogueName(entry.Name())] = entry.Version()
	}

	desired := make([]DesiredArtifact, 0, len(in.Names))
	for _, name := range in.Names {
		version, ok := offered[name]
		switch {
		case !ok:
			result.skip(name, "not offered by the registry")
		case version == "":
			result.skip(name, "source declares no version")
		default:
			desired = append(desired, DesiredArtifact{Kind: in.Kind, Name: name, Version: version})
		}
	}

	return desired, nil
}

// apply commits the planned additions then updates, recording each as added or
// updated and a per-name failure for any fetch or write that fails (FR-006).
func (uc *InstallUseCase) apply(ctx context.Context, run installRun, plan *DiffResponse, result *InstallResponse) {
	for _, want := range plan.Add {
		uc.applyAdd(ctx, run, want, result)
	}
	for _, up := range plan.Update {
		uc.applyUpdate(ctx, run, up, result)
	}
}

// applyAdd fetches, writes, and records one planned fresh artifact, appending it
// to the added group, or recording a per-name failure when the commit fails.
func (uc *InstallUseCase) applyAdd(ctx context.Context, run installRun, want DesiredArtifact, result *InstallResponse) {
	artifact, err := uc.commit(ctx, run, want, "")
	if err != nil {
		result.fail(want.Name, err.Error())
		return
	}

	result.Added = append(result.Added, artifact)
}

// applyUpdate fetches, writes, and records one planned update, preserving the
// prior artifact's InstalledAt from the UpdatePlan, or recording a per-name
// failure when the commit fails.
func (uc *InstallUseCase) applyUpdate(ctx context.Context, run installRun, up UpdatePlan, result *InstallResponse) {
	artifact, err := uc.commit(ctx, run, up.Desired, up.Prior.Spec.InstalledAt)
	if err != nil {
		result.fail(up.Desired.Name, err.Error())
		return
	}

	result.Updated = append(result.Updated, artifact)
}

// commit fetches the artifact's content tree, writes it under the provider
// directory, and records the track document. priorInstalledAt is preserved for an
// update (non-empty) and set to now for a fresh add (empty), bumping updatedAt in
// both cases.
func (uc *InstallUseCase) commit(
	ctx context.Context, run installRun, want DesiredArtifact, priorInstalledAt string,
) (types.Artifact, error) {
	files, err := run.fs.Fetch(ctx, fetchURI(run.kind, want.Name))
	if err != nil {
		return types.Artifact{}, fmt.Errorf("fetch %s: %w", want.Name, err)
	}

	path := installPath(run.kind, want.Name)
	if err := uc.writeTree(ctx, filepath.Join(run.home, path), files); err != nil {
		return types.Artifact{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	installedAt := now
	if priorInstalledAt != "" {
		installedAt = priorInstalledAt
	}
	artifact := types.Artifact{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: run.kind},
		Metadata: types.Metadata{Name: want.Name},
		Spec: types.ArtifactSpec{
			Version:     want.Version,
			Path:        path,
			InstalledAt: installedAt,
			UpdatedAt:   now,
		},
	}
	if err := uc.track.Update(ctx, artifact); err != nil {
		return types.Artifact{}, fmt.Errorf("record %s: %w", want.Name, err)
	}

	return artifact, nil
}

// writeTree writes every file entry under base, recreating the artifact-relative
// directory structure and skipping directory entries.
func (uc *InstallUseCase) writeTree(ctx context.Context, base string, files []source.File) error {
	for _, file := range files {
		if file.IsDirectory() {
			continue
		}
		if err := uc.writeFile(ctx, base, file); err != nil {
			return err
		}
	}

	return nil
}

// writeFile copies one fetched file's content to its artifact-relative location
// under base.
func (uc *InstallUseCase) writeFile(ctx context.Context, base string, file source.File) error {
	dest := filepath.Join(base, file.Name())
	if err := uc.fs.MkdirAll(filepath.Dir(dest), dirPerm); err != nil {
		return fmt.Errorf("prepare %s: %w", dest, err)
	}

	reader, err := file.Read(ctx)
	if err != nil {
		return fmt.Errorf("read %s: %w", file.Name(), err)
	}
	defer func() { _ = reader.Close() }()

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read %s: %w", file.Name(), err)
	}
	if err := afero.WriteFile(uc.fs, dest, body, filePerm); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}

	return nil
}

// installRun bundles the per-invocation context each name is installed against.
type installRun struct {
	kind string
	home string
	fs   source.FileSystem
}
