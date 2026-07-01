package usecase

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// the source roots holding each artifact kind's manifests.
const (
	rootSkills = "skills"
	rootAgents = "agents"
)

// catalogueRoots maps each kind to the source root holding its manifests.
var catalogueRoots = map[CatalogueKind]string{
	CatalogueSkill: rootSkills,
	CatalogueAgent: rootAgents,
}

// ListCatalogueUseCaseParams injects the collaborators the use case composes.
type ListCatalogueUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Open       OpenRegistryUseCase
	Logger     *zap.Logger
}

// ListCatalogueUseCase resolves the single registry, opens its source live, and
// returns the artifact names of the selected kind with the paging window.
type ListCatalogueUseCase struct {
	registries storage.RegistriesStore
	open       OpenRegistryUseCase
	logger     *zap.Logger
}

// NewListCatalogueUseCase builds the use case from the injected collaborators.
func NewListCatalogueUseCase(params ListCatalogueUseCaseParams) *ListCatalogueUseCase {
	return &ListCatalogueUseCase{
		registries: params.Registries,
		open:       params.Open,
		logger:     params.Logger,
	}
}

// Execute runs the validate → get → open → list → collect pipeline, returning a
// classified *Error on the first failing step and otherwise the artifact names
// with their paging window.
func (uc *ListCatalogueUseCase) Execute(ctx context.Context, in ListCatalogueRequest) (*ListCatalogueResponse, error) {
	if err := uc.validate(in); err != nil {
		return nil, err
	}

	registry, err := requireRegistry(ctx, uc.registries)
	if err != nil {
		return nil, err
	}

	fs, err := uc.open.Execute(ctx, *registry)
	if err != nil {
		return nil, err
	}

	files, err := uc.list(ctx, in, fs)
	if err != nil {
		return nil, err
	}

	items := uc.items(files)
	uc.logger.Info("catalogue listed",
		zap.String(telemetry.FieldRegistrySource, registry.Spec.Source),
		zap.Int(telemetry.FieldArtifactCount, len(items)),
	)

	return &ListCatalogueResponse{
		Kind:   in.Kind,
		Items:  items,
		Page:   in.Page,
		Limit:  in.Limit,
		Offset: in.offset(),
	}, nil
}

// validate checks the inputs, returning a usage *Error for any out-of-range
// value. Sort and Order are validated by the handler boundary before Execute,
// so the use case trusts them here.
func (uc *ListCatalogueUseCase) validate(in ListCatalogueRequest) error {
	if _, ok := catalogueRoots[in.Kind]; !ok {
		return NewUsageError(fmt.Sprintf("unknown kind %q", in.Kind))
	}
	if in.Page < 1 {
		return NewUsageError(fmt.Sprintf("page must be at least 1, got %d", in.Page))
	}
	if in.Limit < 1 {
		return NewUsageError(fmt.Sprintf("limit must be at least 1, got %d", in.Limit))
	}

	return nil
}

// list opens the source root for the kind and returns its entries, paging at the
// source with the computed offset.
func (uc *ListCatalogueUseCase) list(ctx context.Context, in ListCatalogueRequest, fs source.FileSystem) ([]source.File, error) {
	opts := []source.Option{
		source.WithSort(in.Sort),
		source.WithOrder(in.Order),
		source.WithOffset(in.offset()),
		source.WithLimit(in.Limit),
	}
	if in.Search != "" {
		opts = append([]source.Option{source.WithSearch(in.Search)}, opts...)
	}

	files, err := fs.List(ctx, catalogueRoots[in.Kind], opts...)
	if err != nil {
		return nil, classifyAdapterErr(err)
	}

	return files, nil
}

// items collects the catalogue names of the manifest entries, skipping directory
// entries; no manifest content is read.
func (uc *ListCatalogueUseCase) items(files []source.File) []string {
	manifests := filterBy(files, func(f source.File) bool { return !f.IsDirectory() })
	names := make([]string, len(manifests))
	for i, f := range manifests {
		names[i] = catalogueName(f.Name())
	}

	return names
}
