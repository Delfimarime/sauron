package usecase

import (
	"context"
	"fmt"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// Lister fetches the page of T selected by the given source.Options — pushed to
// a remote call or applied over an in-memory collection, a decision the adapter
// keeps to itself.
type Lister[T any] interface {
	List(ctx context.Context, opts ...source.Option) ([]T, error)
}

// ListWindow is the search/sort/page window shared by every listing use case's
// request; a request embeds it and it builds the Lister query.
type ListWindow struct {
	Search string
	Sort   string
	Order  string
	Page   int64
	Limit  int64
}

// offset translates the 1-based page and page size to a source offset.
func (w ListWindow) offset() int64 {
	return (w.Page - 1) * w.Limit
}

// options builds the source.Option set a Lister receives for this window.
func (w ListWindow) options() []source.Option {
	opts := []source.Option{
		source.WithSort(w.Sort),
		source.WithOrder(w.Order),
		source.WithOffset(w.offset()),
		source.WithLimit(w.Limit),
	}
	if w.Search != "" {
		opts = append([]source.Option{source.WithSearch(w.Search)}, opts...)
	}
	return opts
}

// validate reports a usage error for an out-of-range page or limit.
func (w ListWindow) validate() error {
	if w.Page < 1 {
		return NewUsageError(fmt.Sprintf("page must be at least 1, got %d", w.Page))
	}
	if w.Limit < 1 {
		return NewUsageError(fmt.Sprintf("limit must be at least 1, got %d", w.Limit))
	}

	return nil
}

// ListResult is the presentation-agnostic outcome shared by every listing use
// case: the fetched page and the paging window applied.
type ListResult[T any] struct {
	Items  []T
	Page   int64
	Limit  int64
	Offset int64
}

// listWith validates window, fetches the page through lister, and wraps it as a
// ListResult. ListUseCase composes this after resolve produces the Lister and
// window for one invocation.
func listWith[T any](ctx context.Context, lister Lister[T], window ListWindow) (*ListResult[T], error) {
	if err := window.validate(); err != nil {
		return nil, err
	}

	items, err := lister.List(ctx, window.options()...)
	if err != nil {
		return nil, err
	}

	return &ListResult[T]{Items: items, Page: window.Page, Limit: window.Limit, Offset: window.offset()}, nil
}

// ListUseCase is the generic listing interactor every listing feature's
// New<Name>UseCase constructor composes (e.g. NewListCatalogueUseCase): given
// the per-invocation input I, resolve produces the Lister[T] bound to its
// source — live remote or local — and the window to fetch; Execute then fetches
// and pages through it. A constructor returns the *ListUseCase[I, T] directly
// when I/T are all its feature's response needs; it wraps one in a dedicated
// type only when the response needs a field the caller doesn't already have.
type ListUseCase[I, T any] struct {
	resolve func(ctx context.Context, in I) (Lister[T], ListWindow, error)
}

// NewListUseCase builds a ListUseCase from resolve, the per-invocation step that
// turns I into a Lister[T] and the window to fetch through it.
func NewListUseCase[I, T any](resolve func(context.Context, I) (Lister[T], ListWindow, error)) *ListUseCase[I, T] {
	return &ListUseCase[I, T]{resolve: resolve}
}

// Execute resolves in to a Lister and window, then fetches and pages through it.
func (uc *ListUseCase[I, T]) Execute(ctx context.Context, in I) (*ListResult[T], error) {
	lister, window, err := uc.resolve(ctx, in)
	if err != nil {
		return nil, err
	}

	return listWith(ctx, lister, window)
}
