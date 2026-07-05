package usecase

import (
	"context"
	"sort"
	"strings"

	"go.uber.org/fx"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const (
	// The fields a tracked-artifact listing may sort by.
	artifactSortName          = "name"
	artifactSortLastUpdatedAt = "lastUpdatedAt"

	// artifactOrderDesc is the descending sort direction; anything else (including
	// empty) sorts ascending.
	artifactOrderDesc = "desc"
)

// ListArtifactsUseCaseParams injects the collaborator the use case composes.
type ListArtifactsUseCaseParams struct {
	fx.In
	Track storage.TrackStore
}

// NewListArtifactsUseCase builds the artifact listing use case on the shared
// listing pipeline: validateKind and validateWindow run before produce builds
// a trackLister bound to the requested kind — the same shape
// NewListCatalogueUseCase's produce closure follows, ending in a local rather
// than a remote Lister. ListArtifactsResponse needs nothing beyond
// ListResult[types.Artifact] (the caller already has Kind before Execute
// runs), so no wrapping type is needed here — the generic Execute is the
// artifact listing use case.
func NewListArtifactsUseCase(params ListArtifactsUseCaseParams) *ListUseCase[ListArtifactsRequest, types.Artifact] {
	return NewPipelineListUseCase(
		func(_ context.Context, in ListArtifactsRequest) (Lister[types.Artifact], error) {
			return trackLister{track: params.Track, kind: in.Kind}, nil
		},
		func(_ context.Context, in ListArtifactsRequest) error { return validateKind(in.Kind) },
		validateWindow[ListArtifactsRequest],
	)
}

// trackLister adapts storage.TrackStore, scoped to one kind, to the shared
// Lister[types.Artifact] seam. track.yaml has no remote pushdown, so List
// resolves the received source.Options into a local source.Options value and
// applies filter, search, sort, and paging itself, entirely in memory, over the
// slice TrackStore.List returns.
type trackLister struct {
	track storage.TrackStore
	kind  string
}

// List reads every tracked artifact, keeps the ones matching kind and (when
// opts carries one) the search term, sorts by the requested field/order, and
// windows at the requested offset/limit.
func (l trackLister) List(ctx context.Context, opts ...source.Option) ([]types.Artifact, error) {
	artifacts, err := l.track.List(ctx)
	if err != nil {
		return nil, ioErr("list track", err)
	}

	var resolved source.Options
	for _, opt := range opts {
		opt(&resolved)
	}

	artifacts = filterBy(artifacts, func(a types.Artifact) bool { return a.Kind == l.kind })
	artifacts = filterBy(artifacts, l.matchesSearch(resolved.Search))
	l.sortArtifacts(artifacts, resolved.Sort, resolved.Order)

	return l.windowArtifacts(artifacts, resolved.Offset, resolved.Limit), nil
}

// matchesSearch builds the predicate keeping artifacts whose name contains term,
// case-insensitive; a nil or empty term keeps every artifact.
func (l trackLister) matchesSearch(term *string) predicate[types.Artifact] {
	if term == nil || *term == "" {
		return func(types.Artifact) bool { return true }
	}

	needle := strings.ToLower(*term)
	return func(a types.Artifact) bool {
		return strings.Contains(strings.ToLower(a.Metadata.Name), needle)
	}
}

// sortArtifacts orders artifacts in place by field ("name", the default, or
// "lastUpdatedAt"), reversing the comparison when order is "desc". Both fields
// are RFC3339 UTC strings or plain names, so a lexicographic string compare
// needs no parsing.
func (l trackLister) sortArtifacts(artifacts []types.Artifact, field, order *string) {
	key := artifactSortName
	if field != nil && *field != "" {
		key = *field
	}
	desc := order != nil && *order == artifactOrderDesc

	sort.SliceStable(artifacts, func(i, j int) bool {
		if desc {
			return l.lessArtifact(artifacts[j], artifacts[i], key)
		}
		return l.lessArtifact(artifacts[i], artifacts[j], key)
	})
}

// lessArtifact compares a and b by the given sort key.
func (l trackLister) lessArtifact(a, b types.Artifact, key string) bool {
	if key == artifactSortLastUpdatedAt {
		return a.Spec.UpdatedAt < b.Spec.UpdatedAt
	}

	return a.Metadata.Name < b.Metadata.Name
}

// windowArtifacts slices items at offset (defaulting to 0), capped to limit
// (defaulting to the remainder); an offset at or past the end yields an empty,
// non-nil slice rather than an error.
func (l trackLister) windowArtifacts(items []types.Artifact, offset, limit *int64) []types.Artifact {
	start := int64(0)
	if offset != nil {
		start = *offset
	}
	if start < 0 || start >= int64(len(items)) {
		return []types.Artifact{}
	}

	end := int64(len(items))
	if limit != nil && start+*limit < end {
		end = start + *limit
	}

	return items[start:end]
}
