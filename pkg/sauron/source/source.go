// Package source defines the read-only filesystem abstraction that registry
// implementations expose once opened.
package source

import (
	"context"
	"errors"
	"io"
)

// ErrNotImplemented is returned by operations that an implementation does not
// yet support.
var ErrNotImplemented = errors.New("not implemented")

// FileSystem is a read-only view over a registry's contents.
type FileSystem interface {
	// List returns the entries under uri, applying any paging given via opts.
	List(ctx context.Context, uri string, opts ...Option) ([]File, error)
	// Describe returns metadata for the entry at uri without reading it.
	Describe(ctx context.Context, uri string) (Stat, error)
	// Get returns the entry at uri.
	Get(ctx context.Context, uri string) (File, error)
}

// File is an entry that exposes its metadata and content.
type File interface {
	Stat
	// Read opens the entry's content. The caller closes the returned reader.
	Read(ctx context.Context) (io.ReadCloser, error)
}

// Stat is the metadata of an entry.
type Stat interface {
	// Name returns the entry's name.
	Name() string
	// IsDirectory reports whether the entry is a directory.
	IsDirectory() bool
	// Size returns the entry's size in bytes.
	Size() int64
	// Version returns the entry's version identifier.
	Version() string
}

// Options configures a listing.
type Options struct {
	Search *string
	Limit  *int64
	Offset *int64
	Sort   *string
}

// Option mutates Options.
type Option func(*Options)

// WithSearch filters the listing by the given term.
func WithSearch(search string) Option {
	return func(o *Options) {
		o.Search = &search
	}
}

// WithLimit caps the number of entries returned.
func WithLimit(limit int64) Option {
	return func(o *Options) {
		o.Limit = &limit
	}
}

// WithOffset skips the given number of entries before returning results.
func WithOffset(offset int64) Option {
	return func(o *Options) {
		o.Offset = &offset
	}
}

// WithSort orders the listing by the given key.
func WithSort(sort string) Option {
	return func(o *Options) {
		o.Sort = &sort
	}
}
