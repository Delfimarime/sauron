package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// catalogueURI is the registry URI reused across the catalogue assertions.
const catalogueURI = "https://acme.example"

// catalogueFixture bundles the use case and its mocked collaborators.
type catalogueFixture struct {
	uc    *ListUseCase[ListCatalogueRequest, string]
	store *storage.MockBasedRegistriesStore
	open  *MockBasedOpenRegistryUseCase
	fs    *source.MockBasedFileSystem
}

// newCatalogueFixture wires the use case over fresh mocks, verifying their
// expectations on cleanup so an unused or over-specified stub fails the test.
func newCatalogueFixture(t *testing.T) *catalogueFixture {
	t.Helper()
	store := &storage.MockBasedRegistriesStore{}
	open := &MockBasedOpenRegistryUseCase{}
	fs := &source.MockBasedFileSystem{}
	t.Cleanup(func() {
		store.AssertExpectations(t)
		open.AssertExpectations(t)
		fs.AssertExpectations(t)
	})
	return &catalogueFixture{
		store: store,
		open:  open,
		fs:    fs,
		uc: NewListCatalogueUseCase(ListCatalogueUseCaseParams{
			Registries: store,
			Open:       open,
			Logger:     zap.NewNop(),
		}),
	}
}

// run executes the use case against the input, returning the result and error.
func (f *catalogueFixture) run(in ListCatalogueRequest) (*ListCatalogueResponse, error) {
	return f.uc.Execute(context.Background(), in)
}

// expectFound stubs Get to return the configured registry.
func (f *catalogueFixture) expectFound() {
	f.store.On("Get", mock.Anything).
		Return(&types.Registry{Spec: types.RegistrySpec{Source: catalogueURI}}, nil)
}

// expectOpen stubs the open action to return the fixture's file system.
func (f *catalogueFixture) expectOpen() {
	f.open.On("Execute", mock.Anything, mock.Anything).Return(f.fs, nil)
}

// expectList stubs List on the given root to return the supplied files,
// capturing the options the use case passed.
func (f *catalogueFixture) expectList(root string, files []source.File, captured *source.Options) {
	f.fs.On("List", mock.Anything, root, mock.Anything).
		Run(func(args mock.Arguments) {
			if captured == nil {
				return
			}
			for _, opt := range args.Get(2).([]source.Option) {
				opt(captured)
			}
		}).
		Return(files, nil)
}

// stat builds a non-directory file stub named name.
func stat(name string) *source.MockBasedFile {
	file := &source.MockBasedFile{}
	file.On("Name").Return(name)
	file.On("IsDirectory").Return(false)
	return file
}

// dir builds a directory entry stub named name.
func dir(name string) *source.MockBasedFile {
	file := &source.MockBasedFile{}
	file.On("Name").Return(name)
	file.On("IsDirectory").Return(true)
	return file
}

// input builds a catalogue input with the sort/order/paging defaults the handler
// boundary resolves before Execute.
func input(kind CatalogueKind) ListCatalogueRequest {
	return ListCatalogueRequest{Kind: kind, ListWindow: ListWindow{Sort: catSortName, Order: catOrderAsc, Page: 1, Limit: 20}}
}

func TestListCatalogueSkillAndAgent(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind CatalogueKind
	}{
		{name: "skill", kind: CatalogueSkill},
		{name: "agent", kind: CatalogueAgent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newCatalogueFixture(t)
			f.expectFound()
			f.expectOpen()
			f.expectList(artifactToDirectoryName[tc.kind], []source.File{
				stat("review.yaml"),
				dir("nested"),
				stat("doc.yml"),
			}, nil)

			// Act.
			result, err := f.run(input(tc.kind))

			// Assert: directories are skipped and names trimmed, for each kind's own root.
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, []string{"review", "doc"}, result.Items)
		})
	}
}

func TestListCatalogueNameTrimming(t *testing.T) {
	// Arrange — .yaml/.yml are trimmed; a name without a known extension is kept.
	f := newCatalogueFixture(t)
	f.expectFound()
	f.expectOpen()
	f.expectList(rootSkills, []source.File{
		stat("alpha.yaml"),
		stat("beta.yml"),
		stat("gamma"),
	}, nil)

	// Act.
	result, err := f.run(input(CatalogueSkill))

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, result.Items)
}

func TestListCatalogueListOptions(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture(t)
	f.expectFound()
	f.expectOpen()
	var captured source.Options
	f.expectList(rootSkills, []source.File{stat("a.yaml")}, &captured)

	in := input(CatalogueSkill)
	in.Search = "rev"
	in.Order = catOrderDesc
	in.Page = 3
	in.Limit = 5

	// Act.
	_, err := f.run(in)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, captured.Search)
	assert.Equal(t, "rev", *captured.Search)
	require.NotNil(t, captured.Sort)
	assert.Equal(t, catSortName, *captured.Sort)
	require.NotNil(t, captured.Order)
	assert.Equal(t, catOrderDesc, *captured.Order)
	require.NotNil(t, captured.Offset)
	assert.Equal(t, int64(10), *captured.Offset) // (3-1)*5
	require.NotNil(t, captured.Limit)
	assert.Equal(t, int64(5), *captured.Limit)
}

func TestListCatalogueNoSearchOption(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture(t)
	f.expectFound()
	f.expectOpen()
	var captured source.Options
	f.expectList(rootAgents, []source.File{stat("a.yaml")}, &captured)

	// Act.
	_, err := f.run(input(CatalogueAgent))

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, captured.Search)
}

func TestListCataloguePagingWindow(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture(t)
	f.expectFound()
	f.expectOpen()
	f.expectList(rootSkills, []source.File{stat("b.yaml")}, nil)

	in := input(CatalogueSkill)
	in.Page = 2
	in.Limit = 1

	// Act.
	result, err := f.run(in)

	// Assert: the result carries the paging window for the client to render.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.Page)
	assert.Equal(t, int64(1), result.Limit)
	assert.Equal(t, int64(1), result.Offset) // (2-1)*1
}

func TestListCatalogueNotConfigured(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture(t)
	f.store.On("Get", mock.Anything).Return(nil, nil)

	// Act.
	_, err := f.run(input(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeNotFound, useErr.Type)
	assert.Contains(t, useErr.Reason, "no registry is configured")
}

func TestListCatalogueReadError(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture(t)
	f.store.On("Get", mock.Anything).Return(nil, errors.New("boom"))

	// Act.
	_, err := f.run(input(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeIO, useErr.Type)
}

func TestListCatalogueUnreachable(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture(t)
	f.expectFound()
	f.open.On("Execute", mock.Anything, mock.Anything).
		Return(nil, NewUnreachableError("source down"))

	// Act.
	_, err := f.run(input(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeUnreachable, useErr.Type)
	assert.Equal(t, "source down", useErr.Reason)
}

func TestListCatalogueListFailureUnreachable(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture(t)
	f.expectFound()
	f.expectOpen()
	f.fs.On("List", mock.Anything, rootSkills, mock.Anything).
		Return(nil, errors.New("connection reset"))

	// Act.
	_, err := f.run(input(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeUnreachable, useErr.Type)
}

func TestListCatalogueUsageErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*ListCatalogueRequest)
		wantSub string
	}{
		{name: "unknown kind", mutate: func(in *ListCatalogueRequest) { in.Kind = "widget" }, wantSub: "kind"},
		{name: "page below one", mutate: func(in *ListCatalogueRequest) { in.Page = 0 }, wantSub: "page"},
		{name: "limit below one", mutate: func(in *ListCatalogueRequest) { in.Limit = 0 }, wantSub: "limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newCatalogueFixture(t)
			in := input(CatalogueSkill)
			tc.mutate(&in)

			// Act.
			_, err := f.run(in)

			// Assert.
			var useErr *Error
			require.ErrorAs(t, err, &useErr)
			assert.Equal(t, TypeUsage, useErr.Type)
			assert.Contains(t, useErr.Reason, tc.wantSub)
			f.store.AssertNotCalled(t, "Get", mock.Anything)
		})
	}
}
