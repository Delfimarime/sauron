package usecase

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// ListCatalogueUseCaseParams injects the collaborators the use case composes.
type ListCatalogueUseCaseParams struct {
	fx.In
	Logger     *zap.Logger
	Registries storage.RegistriesStore
	Open       OpenRegistryUseCase
}

// NewListCatalogueUseCase builds the catalogue listing use case on the shared
// listing pipeline: validateKind and validateWindow run before produce opens
// the registry and builds a catalogueLister bound to the invoked kind's root.
// ListCatalogueResponse needs nothing beyond ListResult[string], so no
// wrapping type is needed here — the generic Execute is the catalogue listing
// use case.
func NewListCatalogueUseCase(params ListCatalogueUseCaseParams) *ListUseCase[ListCatalogueRequest, string] {
	return NewPipelineListUseCase(
		func(ctx context.Context, in ListCatalogueRequest) (Lister[string], error) {
			root := artifactKindDirs[in.Kind] // in.Kind already checked by validateKind

			registry, err := requireRegistry(ctx, params.Registries)
			if err != nil {
				return nil, err
			}

			fs, err := params.Open.Execute(ctx, *registry)
			if err != nil {
				return nil, err
			}

			params.Logger.Info("registry opened", zap.String(telemetry.FieldRegistrySource, registry.Spec.Source))

			return catalogueLister{fs: fs, root: root}, nil
		},
		func(_ context.Context, in ListCatalogueRequest) error { return validateKind(in.Kind) },
		validateWindow[ListCatalogueRequest],
	)
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
