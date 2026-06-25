package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// catalogueReg is the registry name reused across the catalogue assertions.
const catalogueReg = "acme"

// catalogueFixture bundles the use case and its mocked collaborators.
type catalogueFixture struct {
	uc    *ListCatalogueUseCase
	store *storage.MockBasedRegistriesStore
	open  *MockBasedOpenRegistry
	fs    *source.MockBasedFileSystem
}

// newCatalogueFixture wires the use case over fresh mocks.
func newCatalogueFixture() *catalogueFixture {
	store := &storage.MockBasedRegistriesStore{}
	open := &MockBasedOpenRegistry{}
	return &catalogueFixture{
		store: store,
		open:  open,
		fs:    &source.MockBasedFileSystem{},
		uc: NewListCatalogueUseCase(ListCatalogueUseCaseParams{
			Registries: store,
			Open:       open,
			Logger:     zap.NewNop(),
		}),
	}
}

// run executes the use case against the input, returning the result and error.
func (f *catalogueFixture) run(in ListCatalogueInput) (*ListCatalogueResult, error) {
	return f.uc.Execute(context.Background(), in)
}

// names projects a result's entries to their catalogue names.
func names(result *ListCatalogueResult) []string {
	out := make([]string, len(result.Entries))
	for i, entry := range result.Entries {
		out[i] = entry.Name
	}
	return out
}

// expectFound stubs FindByName to return the stored catalogue registry.
func (f *catalogueFixture) expectFound() {
	f.store.On("FindByName", mock.Anything, catalogueReg).
		Return(&types.Registry{Metadata: types.Metadata{Name: catalogueReg}}, nil)
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

// personaFile builds a persona manifest stub whose Read returns content.
func personaFile(name, content string) *source.MockBasedFile {
	file := stat(name)
	file.On("Read", mock.Anything).
		Return(io.NopCloser(strings.NewReader(content)), nil)
	return file
}

// catalogueInput builds a catalogue input with the page/limit defaults applied.
func catalogueInput(kind CatalogueKind) ListCatalogueInput {
	return ListCatalogueInput{
		Kind:     kind,
		Registry: catalogueReg,
		Page:     1,
		Limit:    20,
	}
}

func TestListCatalogueSkillAndAgent(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind CatalogueKind
		root string
	}{
		{name: "skill", kind: CatalogueSkill, root: rootSkills},
		{name: "agent", kind: CatalogueAgent, root: rootAgents},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newCatalogueFixture()
			f.expectFound()
			f.expectOpen()
			f.expectList(tc.root, []source.File{
				stat("review.yaml"),
				dir("nested"),
				stat("doc.yml"),
			}, nil)

			// Act.
			result, err := f.run(catalogueInput(tc.kind))

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, []string{"review", "doc"}, names(result))
			for _, entry := range result.Entries {
				assert.Empty(t, entry.Members)
			}
			assert.Equal(t, int64(1), result.Page)
			assert.Equal(t, int64(20), result.Limit)
		})
	}
}

func TestListCatalogueNameTrimming(t *testing.T) {
	// Arrange — .yaml/.yml are trimmed; a name without a known extension is kept.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	f.expectList(rootSkills, []source.File{
		stat("alpha.yaml"),
		stat("beta.yml"),
		stat("gamma"),
	}, nil)

	// Act.
	result, err := f.run(catalogueInput(CatalogueSkill))

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta", "gamma"}, names(result))
}

func TestListCataloguePersonaMembers(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	f.expectList(rootPersonas, []source.File{
		personaFile("full.yaml", "spec:\n  members:\n    skills: [a, b]\n    agents: [c]\n"),
		personaFile("skillsonly.yaml", "spec:\n  members:\n    skills: [x]\n"),
		personaFile("empty.yaml", "spec:\n  members: {}\n"),
	}, nil)

	// Act.
	result, err := f.run(catalogueInput(CataloguePersona))

	// Assert.
	require.NoError(t, err)
	require.Len(t, result.Entries, 3)
	assert.Equal(t, "skills: a, b; agents: c", result.Entries[0].Members)
	assert.Equal(t, "skills: x", result.Entries[1].Members)
	assert.Equal(t, "—", result.Entries[2].Members)
}

func TestListCatalogueListOptions(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	var captured source.Options
	f.expectList(rootSkills, []source.File{stat("a.yaml")}, &captured)

	in := catalogueInput(CatalogueSkill)
	in.Search = "rev"
	in.Order = orderDesc
	in.Page = 3
	in.Limit = 5

	// Act.
	_, err := f.run(in)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, captured.Search)
	assert.Equal(t, "rev", *captured.Search)
	require.NotNil(t, captured.Sort)
	assert.Equal(t, "name", *captured.Sort)
	require.NotNil(t, captured.Order)
	assert.Equal(t, "desc", *captured.Order)
	require.NotNil(t, captured.Offset)
	assert.Equal(t, int64(10), *captured.Offset) // (3-1)*5
	require.NotNil(t, captured.Limit)
	assert.Equal(t, int64(5), *captured.Limit)
}

func TestListCatalogueNoSearchOption(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	var captured source.Options
	f.expectList(rootAgents, []source.File{stat("a.yaml")}, &captured)

	// Act.
	_, err := f.run(catalogueInput(CatalogueAgent))

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, captured.Search)
}

func TestListCatalogueNotFound(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.store.On("FindByName", mock.Anything, catalogueReg).Return(nil, nil)

	// Act.
	_, err := f.run(catalogueInput(CatalogueSkill))

	// Assert.
	useErr := asUseCaseError(t, err, TypeNotFound)
	assert.Contains(t, useErr.Reason, `registry "acme" does not exist`)
}

func TestListCatalogueReadError(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.store.On("FindByName", mock.Anything, catalogueReg).
		Return(nil, errors.New("boom"))

	// Act.
	_, err := f.run(catalogueInput(CatalogueSkill))

	// Assert.
	_ = asUseCaseError(t, err, TypeIO)
}

func TestListCatalogueUnreachable(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.open.On("Execute", mock.Anything, mock.Anything).
		Return(nil, NewUnreachableError("source down"))

	// Act.
	_, err := f.run(catalogueInput(CatalogueSkill))

	// Assert.
	useErr := asUseCaseError(t, err, TypeUnreachable)
	assert.Equal(t, "source down", useErr.Reason)
}

func TestListCatalogueListFailureUnreachable(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	f.fs.On("List", mock.Anything, rootSkills, mock.Anything).
		Return(nil, errors.New("connection reset"))

	// Act.
	_, err := f.run(catalogueInput(CatalogueSkill))

	// Assert.
	_ = asUseCaseError(t, err, TypeUnreachable)
}

func TestListCataloguePersonaReadError(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	bad := stat("broken.yaml")
	bad.On("Read", mock.Anything).Return(nil, errors.New("io fault"))
	f.expectList(rootPersonas, []source.File{bad}, nil)

	// Act.
	_, err := f.run(catalogueInput(CataloguePersona))

	// Assert.
	_ = asUseCaseError(t, err, TypeIO)
}

func TestListCataloguePersonaDecodeError(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	f.expectList(rootPersonas, []source.File{
		personaFile("bad.yaml", "spec: : not yaml :"),
	}, nil)

	// Act.
	_, err := f.run(catalogueInput(CataloguePersona))

	// Assert.
	_ = asUseCaseError(t, err, TypeIO)
}

func TestListCatalogueUsageErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*ListCatalogueInput)
		wantSub string
	}{
		{name: "unknown kind", mutate: func(in *ListCatalogueInput) { in.Kind = "widget" }, wantSub: "kind"},
		{name: "unknown sort", mutate: func(in *ListCatalogueInput) { in.Sort = "size" }, wantSub: "sort"},
		{name: "unknown order", mutate: func(in *ListCatalogueInput) { in.Order = invalidOrder }, wantSub: "order"},
		{name: "page below one", mutate: func(in *ListCatalogueInput) { in.Page = 0 }, wantSub: "page"},
		{name: "limit below one", mutate: func(in *ListCatalogueInput) { in.Limit = 0 }, wantSub: "limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newCatalogueFixture()
			in := catalogueInput(CatalogueSkill)
			tc.mutate(&in)

			// Act.
			_, err := f.run(in)

			// Assert.
			useErr := asUseCaseError(t, err, TypeUsage)
			assert.Contains(t, useErr.Reason, tc.wantSub)
			f.store.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
		})
	}
}

func TestListCatalogueSortDefaults(t *testing.T) {
	// Arrange — empty Sort/Order default to name/asc and pass validation.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	var captured source.Options
	f.expectList(rootSkills, []source.File{stat("a.yaml")}, &captured)

	in := catalogueInput(CatalogueSkill)
	in.Sort = ""
	in.Order = ""

	// Act.
	_, err := f.run(in)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, captured.Order)
	assert.Equal(t, "asc", *captured.Order)
}
