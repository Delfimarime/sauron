package usecase

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// CatalogueKind is the kind of artifact a catalogue listing browses; it fixes
// the source root listed and the projection applied.
type CatalogueKind string

// The kinds a catalogue listing can browse.
const (
	// CatalogueSkill browses the skills the registry offers under .skills.
	CatalogueSkill CatalogueKind = "skill"
	// CatalogueAgent browses the agents the registry offers under .agents.
	CatalogueAgent CatalogueKind = "agent"
)

// the source roots holding each artifact kind's manifests.
const (
	rootSkills = ".skills"
	rootAgents = ".agents"
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
	Open       OpenRegistry
	Logger     *zap.Logger
}

// ListCatalogueUseCase resolves the single registry, opens its source live, and
// returns the artifact names of the selected kind with the paging window.
type ListCatalogueUseCase struct {
	registries storage.RegistriesStore
	open       OpenRegistry
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
func (uc *ListCatalogueUseCase) Execute(ctx context.Context, in ListCatalogueInput) (*ListCatalogueResult, error) {
	if err := uc.validate(in); err != nil {
		return nil, err
	}

	registry, err := uc.registries.Get(ctx)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read registry: %v", err))
	}
	if registry == nil {
		return nil, NewNotFoundError("no registry is configured")
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
		zap.String(telemetry.FieldRegistryURI, registry.Spec.URI),
		zap.Int(telemetry.FieldArtifactCount, len(items)),
	)

	return &ListCatalogueResult{
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
func (uc *ListCatalogueUseCase) validate(in ListCatalogueInput) error {
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
func (uc *ListCatalogueUseCase) list(ctx context.Context, in ListCatalogueInput, fs source.FileSystem) ([]source.File, error) {
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

// catalogueName trims a manifest's .yaml or .yml extension to its catalogue name.
func catalogueName(filename string) string {
	for _, ext := range []string{".yaml", ".yml"} {
		if strings.HasSuffix(filename, ext) {
			return strings.TrimSuffix(filename, ext)
		}
	}

	return filename
}

// ListCatalogueResult is the presentation-agnostic outcome of browsing the
// catalogue: the artifact names of the kind and the paging window applied.
type ListCatalogueResult struct {
	Kind   CatalogueKind
	Items  []string
	Page   int64
	Limit  int64
	Offset int64
}

// ListCatalogueInput is the per-invocation input for browsing the registry's
// catalogue of one kind.
type ListCatalogueInput struct {
	Kind   CatalogueKind
	Search string
	Sort   string
	Order  string
	Page   int64
	Limit  int64
}

// offset translates the 1-based page and page size to a source offset.
func (in ListCatalogueInput) offset() int64 {
	return (in.Page - 1) * in.Limit
}
