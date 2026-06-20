package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/spf13/afero"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// Directory is a read-only source.FileSystem listing entries under a directory
// tree exposed by an afero.Fs.
type Directory struct {
	fs afero.Fs
}

// NewDirectory builds a Directory over fs.
func NewDirectory(fs afero.Fs) *Directory {
	return &Directory{fs: fs}
}

// List returns the immediate entries under uri, applying any paging given via
// opts.
func (d *Directory) List(_ context.Context, uri string, opts ...source.Option) ([]source.File, error) {
	options := source.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	infos, err := afero.ReadDir(d.fs, uri)
	if err != nil {
		return nil, fmt.Errorf("%w: read %q: %w", ErrRuntime, uri, err)
	}

	return d.page(d.toEntries(infos), options), nil
}

// Describe is not supported by the directory-backed transports yet.
func (d *Directory) Describe(_ context.Context, _ string) (source.Stat, error) {
	return nil, source.ErrNotImplemented
}

// Get is not supported by the directory-backed transports yet.
func (d *Directory) Get(_ context.Context, _ string) (source.File, error) {
	return nil, source.ErrNotImplemented
}

// toEntries maps directory infos to sorted File entries.
func (d *Directory) toEntries(infos []os.FileInfo) []source.File {
	entries := make([]source.File, 0, len(infos))
	for _, info := range infos {
		entries = append(entries, directoryEntry{
			name:  info.Name(),
			isDir: info.IsDir(),
			size:  info.Size(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries
}

// page applies the offset and limit from options to entries.
func (d *Directory) page(entries []source.File, options source.Options) []source.File {
	if options.Offset != nil {
		offset := int(*options.Offset)
		if offset >= len(entries) {
			return []source.File{}
		}
		if offset > 0 {
			entries = entries[offset:]
		}
	}

	if options.Limit != nil {
		limit := int(*options.Limit)
		if limit >= 0 && limit < len(entries) {
			entries = entries[:limit]
		}
	}

	return entries
}

// directoryEntry is a directory listing entry. Its content is not readable yet.
type directoryEntry struct {
	name  string
	isDir bool
	size  int64
}

// Name returns the entry's name.
func (e directoryEntry) Name() string { return e.name }

// IsDirectory reports whether the entry is a directory.
func (e directoryEntry) IsDirectory() bool { return e.isDir }

// Size returns the entry's size in bytes.
func (e directoryEntry) Size() int64 { return e.size }

// Version returns the entry's version identifier, which is not derived yet.
func (e directoryEntry) Version() string { return "" }

// Read is not supported by the directory-backed transports yet.
func (e directoryEntry) Read(_ context.Context) (io.ReadCloser, error) {
	return nil, source.ErrNotImplemented
}
