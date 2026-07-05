package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// Fixture names reused across this file's cases, named once to satisfy goconst.
const (
	skillGoStyle   = "go-style"
	skillSQLReview = "sql-review"
	agentReviewer  = "reviewer"

	artifactAtlas  = "atlas"
	artifactBishop = "bishop"
	artifactCondor = "condor"
)

// artifactsFixture bundles the use case and its mocked collaborator.
type artifactsFixture struct {
	uc    *ListUseCase[ListArtifactsRequest, types.Artifact]
	track *storage.MockBasedTrackStore
}

// newArtifactsFixture wires the use case over a fresh mock, verifying its
// expectations on cleanup so an unused or over-specified stub fails the test.
func newArtifactsFixture(t *testing.T) *artifactsFixture {
	t.Helper()
	track := &storage.MockBasedTrackStore{}
	t.Cleanup(func() {
		track.AssertExpectations(t)
	})
	return &artifactsFixture{
		track: track,
		uc:    NewListArtifactsUseCase(ListArtifactsUseCaseParams{Track: track}),
	}
}

// run executes the use case against the input, returning the result and error.
func (f *artifactsFixture) run(in ListArtifactsRequest) (*ListArtifactsResponse, error) {
	return f.uc.Execute(context.Background(), in)
}

// expectList stubs TrackStore.List to return the supplied artifacts.
func (f *artifactsFixture) expectList(artifacts []types.Artifact, err error) {
	f.track.On("List", mock.Anything).Return(artifacts, err)
}

// artifact builds a tracked Artifact stub with the given kind, name, version,
// and lastUpdatedAt timestamp.
func artifact(kind, name, version, updatedAt string) types.Artifact {
	return types.Artifact{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: kind},
		Metadata: types.Metadata{Name: name},
		Spec:     types.ArtifactSpec{Version: version, UpdatedAt: updatedAt},
	}
}

// artifactsInput builds a request with the sort/order/paging defaults the
// handler boundary resolves before Execute.
func artifactsInput(kind string) ListArtifactsRequest {
	return ListArtifactsRequest{
		Kind:       kind,
		ListWindow: ListWindow{Sort: artifactSortName, Order: catOrderAsc, Page: 1, Limit: 20},
	}
}

// names extracts the Metadata.Name of each item, in order, for assertion.
func names(items []types.Artifact) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Metadata.Name
	}

	return result
}

func TestListArtifactsFiltersByKind(t *testing.T) {
	for _, tc := range []struct {
		kind string
		want []string
	}{
		{kind: types.KindSkill, want: []string{skillGoStyle, skillSQLReview}},
		{kind: types.KindAgent, want: []string{agentReviewer}},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			// Arrange.
			f := newArtifactsFixture(t)
			f.expectList([]types.Artifact{
				artifact(types.KindSkill, skillGoStyle, "v1", "2026-06-15T00:00:00Z"),
				artifact(types.KindAgent, agentReviewer, "v2", "2026-06-14T00:00:00Z"),
				artifact(types.KindSkill, skillSQLReview, "", "2026-06-12T00:00:00Z"),
			}, nil)

			// Act.
			result, err := f.run(artifactsInput(tc.kind))

			// Assert.
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.want, names(result.Items))
		})
	}
}

func TestListArtifactsSearchFilter(t *testing.T) {
	// Arrange — the search term matches case-insensitively, on name only.
	f := newArtifactsFixture(t)
	f.expectList([]types.Artifact{
		artifact(types.KindSkill, "Go Style", "v1", "2026-06-15T00:00:00Z"),
		artifact(types.KindSkill, skillSQLReview, "", "2026-06-12T00:00:00Z"),
	}, nil)

	in := artifactsInput(types.KindSkill)
	in.Search = "GO"

	// Act.
	result, err := f.run(in)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, []string{"Go Style"}, names(result.Items))
}

func TestListArtifactsSort(t *testing.T) {
	tests := []struct {
		name  string
		sort  string
		order string
		want  []string
	}{
		{
			name: "name asc", sort: artifactSortName, order: catOrderAsc,
			want: []string{artifactAtlas, artifactBishop, artifactCondor},
		},
		{
			name: "name desc", sort: artifactSortName, order: catOrderDesc,
			want: []string{artifactCondor, artifactBishop, artifactAtlas},
		},
		{
			name: "lastUpdatedAt asc", sort: artifactSortLastUpdatedAt, order: catOrderAsc,
			want: []string{artifactCondor, artifactAtlas, artifactBishop},
		},
		{
			name: "lastUpdatedAt desc", sort: artifactSortLastUpdatedAt, order: catOrderDesc,
			want: []string{artifactBishop, artifactAtlas, artifactCondor},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newArtifactsFixture(t)
			f.expectList([]types.Artifact{
				artifact(types.KindSkill, artifactBishop, "v1", "2026-06-20T00:00:00Z"),
				artifact(types.KindSkill, artifactAtlas, "v1", "2026-06-15T00:00:00Z"),
				artifact(types.KindSkill, artifactCondor, "v1", "2026-06-10T00:00:00Z"),
			}, nil)

			in := artifactsInput(types.KindSkill)
			in.Sort = tc.sort
			in.Order = tc.order

			// Act.
			result, err := f.run(in)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tc.want, names(result.Items))
		})
	}
}

func TestListArtifactsPagingWindow(t *testing.T) {
	// Arrange.
	f := newArtifactsFixture(t)
	f.expectList([]types.Artifact{
		artifact(types.KindSkill, artifactAtlas, "v1", "2026-06-10T00:00:00Z"),
		artifact(types.KindSkill, artifactBishop, "v1", "2026-06-11T00:00:00Z"),
		artifact(types.KindSkill, artifactCondor, "v1", "2026-06-12T00:00:00Z"),
	}, nil)

	in := artifactsInput(types.KindSkill)
	in.Page = 2
	in.Limit = 1

	// Act.
	result, err := f.run(in)

	// Assert: the middle item, and the applied window is reported back.
	require.NoError(t, err)
	assert.Equal(t, []string{artifactBishop}, names(result.Items))
	assert.Equal(t, int64(2), result.Page)
	assert.Equal(t, int64(1), result.Limit)
	assert.Equal(t, int64(1), result.Offset) // (2-1)*1
}

func TestListArtifactsPageBeyondEnd(t *testing.T) {
	// Arrange.
	f := newArtifactsFixture(t)
	f.expectList([]types.Artifact{
		artifact(types.KindSkill, artifactAtlas, "v1", "2026-06-10T00:00:00Z"),
	}, nil)

	in := artifactsInput(types.KindSkill)
	in.Page = 5
	in.Limit = 20

	// Act.
	result, err := f.run(in)

	// Assert: an empty window, no error.
	require.NoError(t, err)
	assert.Empty(t, result.Items)
}

func TestListArtifactsEmptyKind(t *testing.T) {
	// Arrange — nothing tracked of the requested kind.
	f := newArtifactsFixture(t)
	f.expectList([]types.Artifact{
		artifact(types.KindAgent, agentReviewer, "v1", "2026-06-10T00:00:00Z"),
	}, nil)

	// Act.
	result, err := f.run(artifactsInput(types.KindSkill))

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, result.Items)
}

func TestListArtifactsReadError(t *testing.T) {
	// Arrange.
	f := newArtifactsFixture(t)
	f.expectList(nil, errors.New("boom"))

	// Act.
	_, err := f.run(artifactsInput(types.KindSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeIO, useErr.Type)
}

func TestListArtifactsUsageErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*ListArtifactsRequest)
		wantSub string
	}{
		{name: "unknown kind", mutate: func(in *ListArtifactsRequest) { in.Kind = "Widget" }, wantSub: "kind"},
		{name: "page below one", mutate: func(in *ListArtifactsRequest) { in.Page = 0 }, wantSub: "page"},
		{name: "limit below one", mutate: func(in *ListArtifactsRequest) { in.Limit = 0 }, wantSub: "limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newArtifactsFixture(t)
			in := artifactsInput(types.KindSkill)
			tc.mutate(&in)

			// Act.
			_, err := f.run(in)

			// Assert.
			var useErr *Error
			require.ErrorAs(t, err, &useErr)
			assert.Equal(t, TypeUsage, useErr.Type)
			assert.Contains(t, useErr.Reason, tc.wantSub)
			f.track.AssertNotCalled(t, "List", mock.Anything)
		})
	}
}
