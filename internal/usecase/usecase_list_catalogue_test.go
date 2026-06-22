package usecase

import (
	"bytes"
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

// run executes the use case against the request, returning the output and error.
func (f *catalogueFixture) run(request *ListCatalogueRequest) (string, error) {
	var out bytes.Buffer
	request.out = &out
	err := f.uc.Execute(request)
	return out.String(), err
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

// request builds a catalogue request with the page/limit defaults applied.
func request(kind CatalogueKind) *ListCatalogueRequest {
	req := NewListCatalogueRequest(context.Background(), nil)
	req.Kind = kind
	req.Registry = catalogueReg
	req.Page = 1
	req.Limit = 20
	return req
}

// lines splits rendered output into its non-empty trimmed lines.
func lines(out string) []string {
	return strings.Split(strings.TrimRight(out, "\n"), "\n")
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
			out, err := f.run(request(tc.kind))

			// Assert.
			require.NoError(t, err)
			rows := lines(out)
			assert.Equal(t, "NAME", strings.Fields(rows[0])[0])
			assert.Equal(t, "KIND", strings.Fields(rows[0])[1])
			assert.Contains(t, out, "review")
			assert.Contains(t, out, "doc")
			assert.NotContains(t, out, "nested")
			assert.Contains(t, out, tc.name)
			assert.Contains(t, out, "showing 1–2 (page 1, limit 20)")
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
	out, err := f.run(request(CatalogueSkill))

	// Assert.
	require.NoError(t, err)
	assert.Contains(t, out, "alpha")
	assert.Contains(t, out, "beta")
	assert.Contains(t, out, "gamma")
	assert.NotContains(t, out, "alpha.yaml")
	assert.NotContains(t, out, "beta.yml")
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
	out, err := f.run(request(CataloguePersona))

	// Assert.
	require.NoError(t, err)
	assert.Contains(t, lines(out)[0], "MEMBERS")
	assert.Contains(t, out, "skills: a, b; agents: c")
	assert.Contains(t, out, "skills: x")
	assert.Contains(t, out, "—")
	assert.Contains(t, out, "showing 1–3 (page 1, limit 20)")
}

func TestListCatalogueListOptions(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.expectOpen()
	var captured source.Options
	f.expectList(rootSkills, []source.File{stat("a.yaml")}, &captured)

	req := request(CatalogueSkill)
	req.Search = "rev"
	req.Order = orderDesc
	req.Page = 3
	req.Limit = 5

	// Act.
	_, err := f.run(req)

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
	_, err := f.run(request(CatalogueAgent))

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, captured.Search)
}

func TestListCataloguePagingLine(t *testing.T) {
	for _, tc := range []struct {
		name  string
		page  int64
		limit int64
		files []source.File
		want  string
	}{
		{
			name:  "populated window",
			page:  2,
			limit: 1,
			files: []source.File{stat("b.yaml")},
			want:  "showing 2–2 (page 2, limit 1)",
		},
		{
			name:  "empty page",
			page:  9,
			limit: 20,
			files: nil,
			want:  "showing 0 results (page 9, limit 20)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newCatalogueFixture()
			f.expectFound()
			f.expectOpen()
			f.expectList(rootSkills, tc.files, nil)

			req := request(CatalogueSkill)
			req.Page = tc.page
			req.Limit = tc.limit

			// Act.
			out, err := f.run(req)

			// Assert.
			require.NoError(t, err)
			assert.Contains(t, out, tc.want)
		})
	}
}

func TestListCatalogueNotFound(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.store.On("FindByName", mock.Anything, catalogueReg).Return(nil, nil)

	// Act.
	_, err := f.run(request(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeNotFound, useErr.Type)
	assert.Contains(t, useErr.Reason, `registry "acme" does not exist`)
}

func TestListCatalogueReadError(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.store.On("FindByName", mock.Anything, catalogueReg).
		Return(nil, errors.New("boom"))

	// Act.
	_, err := f.run(request(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeIO, useErr.Type)
}

func TestListCatalogueUnreachable(t *testing.T) {
	// Arrange.
	f := newCatalogueFixture()
	f.expectFound()
	f.open.On("Execute", mock.Anything, mock.Anything).
		Return(nil, NewUnreachableError("source down"))

	// Act.
	_, err := f.run(request(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeUnreachable, useErr.Type)
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
	_, err := f.run(request(CatalogueSkill))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeUnreachable, useErr.Type)
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
	_, err := f.run(request(CataloguePersona))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeIO, useErr.Type)
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
	_, err := f.run(request(CataloguePersona))

	// Assert.
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.Equal(t, TypeIO, useErr.Type)
}

func TestListCatalogueUsageErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*ListCatalogueRequest)
		wantSub string
	}{
		{name: "unknown kind", mutate: func(r *ListCatalogueRequest) { r.Kind = "widget" }, wantSub: "kind"},
		{name: "unknown sort", mutate: func(r *ListCatalogueRequest) { r.Sort = "size" }, wantSub: "sort"},
		{name: "unknown order", mutate: func(r *ListCatalogueRequest) { r.Order = invalidOrder }, wantSub: "order"},
		{name: "page below one", mutate: func(r *ListCatalogueRequest) { r.Page = 0 }, wantSub: "page"},
		{name: "limit below one", mutate: func(r *ListCatalogueRequest) { r.Limit = 0 }, wantSub: "limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			f := newCatalogueFixture()
			req := request(CatalogueSkill)
			tc.mutate(req)

			// Act.
			_, err := f.run(req)

			// Assert.
			var useErr *Error
			require.ErrorAs(t, err, &useErr)
			assert.Equal(t, TypeUsage, useErr.Type)
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

	req := request(CatalogueSkill)
	req.Sort = ""
	req.Order = ""

	// Act.
	_, err := f.run(req)

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, captured.Order)
	assert.Equal(t, "asc", *captured.Order)
}
