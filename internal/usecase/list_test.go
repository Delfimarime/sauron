package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// fakeLister is a minimal Lister[T] test double that records the resolved
// source.Options it received and returns a fixed items/error pair.
type fakeLister[T any] struct {
	items   []T
	err     error
	options source.Options
}

func (f *fakeLister[T]) List(_ context.Context, opts ...source.Option) ([]T, error) {
	for _, opt := range opts {
		opt(&f.options)
	}
	return f.items, f.err
}

func TestListWith(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// window is the request-side search/sort/page input.
		window ListWindow
		// lister is the fake Lister[string] the call is composed over.
		lister *fakeLister[string]
		// wantErr is a substring the error must contain; empty means no error.
		wantErr string
		// wantResult is the expected outcome when no error is wanted.
		wantResult *ListResult[string]
		// wantSearchOmitted asserts the resolved options carry no search term.
		wantSearchOmitted bool
	}{
		{
			name:       "search/sort/order/page/limit resolve into the fetched window",
			window:     ListWindow{Search: "go", Sort: artifactSortName, Order: "asc", Page: 2, Limit: 10},
			lister:     &fakeLister[string]{items: []string{"a", "b"}},
			wantResult: &ListResult[string]{Items: []string{"a", "b"}, Page: 2, Limit: 10, Offset: 10},
		},
		{
			name:              "empty search omits the search option",
			window:            ListWindow{Sort: "name", Order: "asc", Page: 1, Limit: 20},
			lister:            &fakeLister[string]{items: nil},
			wantResult:        &ListResult[string]{Items: nil, Page: 1, Limit: 20, Offset: 0},
			wantSearchOmitted: true,
		},
		{
			name:    "a lister failure is returned unchanged",
			window:  ListWindow{Page: 1, Limit: 20},
			lister:  &fakeLister[string]{err: NewUnreachableError("boom")},
			wantErr: "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			result, err := listWith[string](context.Background(), tt.lister, tt.window)

			// Assert.
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
			if tt.wantSearchOmitted {
				assert.Nil(t, tt.lister.options.Search)
			}
		})
	}
}

// TestListUseCaseExecute exercises the generic ListUseCase: Execute resolves I to
// a Lister and window via resolve, then fetches and pages through it — the shape
// every listing feature's New<Name>UseCase constructor (e.g.
// NewListCatalogueUseCase) composes.
func TestListUseCaseExecute(t *testing.T) {
	t.Run("delegates to the resolved lister and window", func(t *testing.T) {
		// Arrange.
		lister := &fakeLister[string]{items: []string{"a", "b"}}
		uc := NewListUseCase(func(_ context.Context, _ string) (Lister[string], ListWindow, error) {
			return lister, ListWindow{Page: 1, Limit: 20}, nil
		})

		// Act.
		result, err := uc.Execute(context.Background(), "anything")

		// Assert.
		require.NoError(t, err)
		assert.Equal(t, &ListResult[string]{Items: []string{"a", "b"}, Page: 1, Limit: 20, Offset: 0}, result)
	})

	t.Run("a resolve failure is returned unchanged, never reaching the lister", func(t *testing.T) {
		// Arrange.
		uc := NewListUseCase(func(_ context.Context, _ string) (Lister[string], ListWindow, error) {
			return nil, ListWindow{}, NewUsageError("boom")
		})

		// Act.
		_, err := uc.Execute(context.Background(), "anything")

		// Assert.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
	})
}

// TestListWindowOptions confirms every populated field resolves to its matching
// source.Option, including the offset computed from page and limit.
func TestListWindowOptions(t *testing.T) {
	window := ListWindow{Search: "go", Sort: "name", Order: "desc", Page: 3, Limit: 5}

	var resolved source.Options
	for _, opt := range window.options() {
		opt(&resolved)
	}

	require.NotNil(t, resolved.Search)
	assert.Equal(t, "go", *resolved.Search)
	require.NotNil(t, resolved.Sort)
	assert.Equal(t, "name", *resolved.Sort)
	require.NotNil(t, resolved.Order)
	assert.Equal(t, "desc", *resolved.Order)
	require.NotNil(t, resolved.Offset)
	assert.Equal(t, int64(10), *resolved.Offset)
	require.NotNil(t, resolved.Limit)
	assert.Equal(t, int64(5), *resolved.Limit)
}

// TestListWindowValidate asserts ListWindow.validate rejects an out-of-range
// page or limit and accepts an in-range window.
func TestListWindowValidate(t *testing.T) {
	tests := []struct {
		name    string
		window  ListWindow
		wantErr string
	}{
		{name: "page below 1 is a usage error", window: ListWindow{Page: 0, Limit: 20}, wantErr: "page must be at least 1, got 0"},
		{name: "limit below 1 is a usage error", window: ListWindow{Page: 1, Limit: 0}, wantErr: "limit must be at least 1, got 0"},
		{name: "an in-range window is valid", window: ListWindow{Page: 1, Limit: 20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.window.validate()

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

// pipelineRequest is a minimal hasWindow request used to exercise
// NewPipelineListUseCase and validateWindow independent of any real feature.
type pipelineRequest struct {
	ListWindow
}

// TestListWindowWindowMethod asserts Window returns the embedded value
// unchanged, so an embedding request satisfies hasWindow via promotion.
func TestListWindowWindowMethod(t *testing.T) {
	window := ListWindow{Page: 2, Limit: 10}
	req := pipelineRequest{ListWindow: window}

	assert.Equal(t, window, req.Window())
}

// TestValidateWindow asserts validateWindow delegates to the request's
// embedded ListWindow.validate.
func TestValidateWindow(t *testing.T) {
	err := validateWindow(context.Background(), pipelineRequest{ListWindow: ListWindow{Page: 0, Limit: 20}})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "page must be at least 1")

	require.NoError(t, validateWindow(context.Background(), pipelineRequest{ListWindow: ListWindow{Page: 1, Limit: 20}}))
}

// TestNewPipelineListUseCase exercises the Chain of Responsibility: a
// validator halts the chain before produce ever runs, and once every
// validator passes, produce's Lister and the request's window reach Execute.
func TestNewPipelineListUseCase(t *testing.T) {
	t.Run("an invalid window is rejected before produce runs", func(t *testing.T) {
		// Arrange.
		produceCalled := false
		uc := NewPipelineListUseCase(
			func(_ context.Context, _ pipelineRequest) (Lister[string], error) {
				produceCalled = true
				return &fakeLister[string]{}, nil
			},
			validateWindow[pipelineRequest],
		)

		// Act.
		_, err := uc.Execute(context.Background(), pipelineRequest{ListWindow: ListWindow{Page: 0, Limit: 20}})

		// Assert.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "page must be at least 1")
		assert.False(t, produceCalled, "produce must not run once a validator fails")
	})

	t.Run("a failing validator earlier in the chain halts before a later one runs", func(t *testing.T) {
		// Arrange.
		laterRan := false
		uc := NewPipelineListUseCase(
			func(_ context.Context, _ pipelineRequest) (Lister[string], error) {
				return &fakeLister[string]{}, nil
			},
			func(_ context.Context, _ pipelineRequest) error { return NewUsageError("first validator failed") },
			func(_ context.Context, _ pipelineRequest) error { laterRan = true; return nil },
		)

		// Act.
		_, err := uc.Execute(context.Background(), pipelineRequest{ListWindow: ListWindow{Page: 1, Limit: 20}})

		// Assert.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "first validator failed")
		assert.False(t, laterRan, "a later validator must not run once an earlier one fails")
	})

	t.Run("once every validator passes, produce's lister is fetched through the request's window", func(t *testing.T) {
		// Arrange.
		lister := &fakeLister[string]{items: []string{"a", "b"}}
		uc := NewPipelineListUseCase(
			func(_ context.Context, _ pipelineRequest) (Lister[string], error) {
				return lister, nil
			},
			validateWindow[pipelineRequest],
		)

		// Act.
		result, err := uc.Execute(context.Background(), pipelineRequest{ListWindow: ListWindow{Page: 2, Limit: 10}})

		// Assert.
		require.NoError(t, err)
		assert.Equal(t, &ListResult[string]{Items: []string{"a", "b"}, Page: 2, Limit: 10, Offset: 10}, result)
	})

	t.Run("a produce failure is returned unchanged", func(t *testing.T) {
		// Arrange.
		uc := NewPipelineListUseCase(
			func(_ context.Context, _ pipelineRequest) (Lister[string], error) {
				return nil, NewUnreachableError("source down")
			},
			validateWindow[pipelineRequest],
		)

		// Act.
		_, err := uc.Execute(context.Background(), pipelineRequest{ListWindow: ListWindow{Page: 1, Limit: 20}})

		// Assert.
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source down")
	})
}
