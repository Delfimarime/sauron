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

var (
	artifactToDirectoryName = map[CatalogueKind]string{CatalogueSkill: rootSkills, CatalogueAgent: rootAgents}
)

// ListCatalogueUseCaseParams injects the collaborators the use case composes.
type ListCatalogueUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Open       OpenRegistryUseCase
	Logger     *zap.Logger
}

// NewListCatalogueUseCase builds the catalogue listing use case: the generic
// ListUseCase, whose resolve step is the validate → get → open pipeline (what
// used to be a bespoke Execute body) ending in a catalogueLister bound to the
// invoked kind's root. ListCatalogueResponse needs nothing beyond
// ListResult[string], so no wrapping type is needed here — the generic Execute
// is the catalogue listing use case.
func NewListCatalogueUseCase(params ListCatalogueUseCaseParams) *ListUseCase[ListCatalogueRequest, string] {
	return NewListUseCase(func(ctx context.Context, in ListCatalogueRequest) (Lister[string], ListWindow, error) {
		root, ok := artifactToDirectoryName[in.Kind]
		if !ok {
			return nil, ListWindow{}, NewUsageError(fmt.Sprintf("unknown kind %q", in.Kind))
		}
		if err := in.ListWindow.validate(); err != nil {
			return nil, ListWindow{}, err
		}

		registry, err := requireRegistry(ctx, params.Registries)
		if err != nil {
			return nil, ListWindow{}, err
		}

		fs, err := params.Open.Execute(ctx, *registry)
		if err != nil {
			return nil, ListWindow{}, err
		}

		params.Logger.Info("registry opened", zap.String(telemetry.FieldRegistrySource, registry.Spec.Source))

		return catalogueLister{fs: fs, root: root}, in.ListWindow, nil
	})
}

// catalogueLister adapts an opened registry source.FileSystem, scoped to one
// kind's root, to the shared Lister[string] seam: it pushes the window's
// options to the remote List call and projects each manifest entry to its
// catalogue name, skipping directories.
type catalogueLister struct {
	fs   source.FileSystem
	root string
}

// List fetches root's entries through fs, applying opts remotely, and returns
// the non-directory entries' trimmed catalogue names.
func (l catalogueLister) List(ctx context.Context, opts ...source.Option) ([]string, error) {
	files, err := l.fs.List(ctx, l.root, opts...)
	if err != nil {
		return nil, classifyAdapterErr(err)
	}

	manifests := filterBy(files, func(f source.File) bool { return !f.IsDirectory() })
	names := make([]string, len(manifests))
	for i, f := range manifests {
		names[i] = catalogueName(f.Name())
	}

	return names, nil
}
