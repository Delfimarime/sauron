package usecase

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// altName is the second registry name reused across the listing assertions;
// headName is the repeated name-column header.
const (
	altName  = "internal"
	headName = "NAME"
)

// reg builds a Registry with the given name, transport, and uri.
func reg(name string, transport types.Transport, uri string) types.Registry {
	return types.Registry{
		Metadata: types.Metadata{Name: name},
		Spec:     types.RegistrySpec{Transport: transport, URI: uri},
	}
}

// listFixture bundles the list use case and its mocked store.
type listFixture struct {
	uc    *ListRegistriesUseCase
	store *storage.MockBasedRegistriesStore
}

// newListFixture wires a list use case over a fresh store mock.
func newListFixture() *listFixture {
	store := &storage.MockBasedRegistriesStore{}
	return &listFixture{
		store: store,
		uc: NewListRegistriesUseCase(ListRegistriesUseCaseParams{
			Registries: store,
			Logger:     zap.NewNop(),
		}),
	}
}

// run executes the use case against the request, returning the output and error.
func (f *listFixture) run(request *ListRegistriesRequest, out *bytes.Buffer) error {
	request.out = out
	return f.uc.Execute(request)
}

// dataRows splits a rendered table into its non-header rows.
func dataRows(out string) []string {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) <= 1 {
		return nil
	}
	return lines[1:]
}

// TestListRegistriesSuccess covers the read → filter → sort → project → render
// pipeline across the field, search, sort, and order axes.
func TestListRegistriesSuccess(t *testing.T) {
	stored := []types.Registry{
		reg(altName, types.TransportHTTP, "https://reg.example.com/"),
		reg(testName, types.TransportGit, "git@github.com:acme/artifacts.git"),
	}

	tests := []struct {
		name      string
		stored    []types.Registry
		request   ListRegistriesRequest
		wantHead  []string
		wantOrder []string
		wantEmpty bool
	}{
		{
			name:      "default columns and sort by name asc",
			stored:    stored,
			request:   ListRegistriesRequest{},
			wantHead:  []string{headName, "TRANSPORT", "URI"},
			wantOrder: []string{testName, altName},
		},
		{
			name:      "fields selects and forces name first",
			stored:    stored,
			request:   ListRegistriesRequest{Fields: []string{"uri"}},
			wantHead:  []string{headName, "URI"},
			wantOrder: []string{testName, altName},
		},
		{
			name:      "every column including the optionals",
			stored:    stored,
			request:   ListRegistriesRequest{Fields: []string{fieldName, fieldTransport, fieldURI, fieldRef, fieldTimeout}},
			wantOrder: []string{testName, altName},
		},
		{
			name:      "case-insensitive search keeps the match",
			stored:    stored,
			request:   ListRegistriesRequest{Search: "ACME"},
			wantOrder: []string{testName},
		},
		{
			name:      "sort transport order desc",
			stored:    stored,
			request:   ListRegistriesRequest{Sort: "transport", Order: "desc"},
			wantOrder: []string{altName, testName},
		},
		{
			name: "transport tie breaks on name asc",
			stored: []types.Registry{
				reg("zeta", types.TransportHTTP, "https://z/"),
				reg("alpha", types.TransportHTTP, "https://a/"),
			},
			request:   ListRegistriesRequest{Sort: "transport"},
			wantOrder: []string{"alpha", "zeta"},
		},
		{
			name:      "empty result renders nothing",
			stored:    nil,
			request:   ListRegistriesRequest{},
			wantEmpty: true,
		},
		{
			name:      "search with no match renders nothing",
			stored:    stored,
			request:   ListRegistriesRequest{Search: "absent"},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			f := newListFixture()
			f.store.On("List", mock.Anything).Return(tt.stored, nil)
			request := tt.request
			request.Context = context.Background()
			var out bytes.Buffer

			// Act.
			err := f.run(&request, &out)

			// Assert.
			require.NoError(t, err)
			if tt.wantEmpty {
				assert.Empty(t, out.String())
				return
			}
			lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
			if tt.wantHead != nil {
				for _, h := range tt.wantHead {
					assert.Contains(t, lines[0], h)
				}
			}
			rows := dataRows(out.String())
			require.Len(t, rows, len(tt.wantOrder))
			for i, want := range tt.wantOrder {
				assert.Truef(t, strings.HasPrefix(rows[i], want),
					"row %d %q starts with %q", i, rows[i], want)
			}
		})
	}
}

// TestListRegistriesRendersEmptyCell asserts an absent optional column renders
// the placeholder dash.
func TestListRegistriesRendersEmptyCell(t *testing.T) {
	// Arrange.
	f := newListFixture()
	f.store.On("List", mock.Anything).
		Return([]types.Registry{reg(testName, types.TransportHTTP, "https://a/")}, nil)
	request := ListRegistriesRequest{Context: context.Background(), Fields: []string{"ref"}}
	var out bytes.Buffer

	// Act.
	err := f.run(&request, &out)

	// Assert.
	require.NoError(t, err)
	assert.Contains(t, out.String(), "—")
}

// TestListRegistriesUsageErrors covers the out-of-set flag values that classify
// as usage failures.
func TestListRegistriesUsageErrors(t *testing.T) {
	tests := []struct {
		name    string
		request ListRegistriesRequest
	}{
		{name: "unknown field", request: ListRegistriesRequest{Fields: []string{"bogus"}}},
		{name: "unknown sort", request: ListRegistriesRequest{Sort: "uri"}},
		{name: "unknown order", request: ListRegistriesRequest{Order: "sideways"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			f := newListFixture()
			request := tt.request
			request.Context = context.Background()
			var out bytes.Buffer

			// Act.
			err := f.run(&request, &out)

			// Assert: classified usage, store never consulted, no output.
			_ = asUseCaseError(t, err, TypeUsage)
			assert.Empty(t, out.String())
			f.store.AssertNotCalled(t, "List", mock.Anything)
		})
	}
}

// TestListRegistriesIOError asserts a failing read classifies as an io failure.
func TestListRegistriesIOError(t *testing.T) {
	// Arrange.
	f := newListFixture()
	f.store.On("List", mock.Anything).Return(nil, errors.New("registries.yaml is unreadable"))
	request := ListRegistriesRequest{Context: context.Background()}
	var out bytes.Buffer

	// Act.
	err := f.run(&request, &out)

	// Assert.
	_ = asUseCaseError(t, err, TypeIO)
	assert.Empty(t, out.String())
}

// TestNewListRegistriesRequest asserts the constructor binds the context and
// output writer.
func TestNewListRegistriesRequest(t *testing.T) {
	// Arrange.
	var out bytes.Buffer

	// Act.
	request := NewListRegistriesRequest(context.Background(), &out)

	// Assert.
	require.NotNil(t, request)
	assert.Same(t, &out, request.Out())
}
