package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

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

	return d.page(d.toEntries(d.filter(infos, options), uri, options), options), nil
}

// filter keeps only the infos whose name contains options.Search, comparing
// case-insensitively; a nil or empty Search matches every info.
func (d *Directory) filter(infos []os.FileInfo, options source.Options) []os.FileInfo {
	if options.Search == nil || *options.Search == "" {
		return infos
	}

	term := strings.ToLower(*options.Search)
	matched := make([]os.FileInfo, 0, len(infos))
	for _, info := range infos {
		if strings.Contains(strings.ToLower(info.Name()), term) {
			matched = append(matched, info)
		}
	}

	return matched
}

// Describe is not supported by the directory-backed transports yet.
func (d *Directory) Describe(_ context.Context, _ string) (source.Stat, error) {
	return nil, source.ErrNotImplemented
}

// Get is not supported by the directory-backed transports yet.
func (d *Directory) Get(_ context.Context, _ string) (source.File, error) {
	return nil, source.ErrNotImplemented
}

// toEntries maps directory infos to File entries sorted by name in the direction
// given by options; the order is ascending unless Order is "desc".
func (d *Directory) toEntries(infos []os.FileInfo, uri string, options source.Options) []source.File {
	entries := make([]source.File, 0, len(infos))
	for _, info := range infos {
		entries = append(entries, directoryEntry{
			fs:    d.fs,
			path:  path.Join(uri, info.Name()),
			name:  info.Name(),
			isDir: info.IsDir(),
			size:  info.Size(),
		})
	}

	descending := options.Order != nil && *options.Order == "desc"
	sort.Slice(entries, func(i, j int) bool {
		if descending {
			return entries[i].Name() > entries[j].Name()
		}
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

// directoryEntry is a directory listing entry that reads its content through fs.
type directoryEntry struct {
	fs    afero.Fs
	path  string
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

// Read opens the entry's content through its filesystem; the caller closes the
// returned reader. Directories are not readable.
func (e directoryEntry) Read(_ context.Context) (io.ReadCloser, error) {
	if e.isDir {
		return nil, fmt.Errorf("%w: read %q: is a directory", ErrRuntime, e.path)
	}

	file, err := e.fs.Open(e.path)
	if err != nil {
		return nil, fmt.Errorf("%w: read %q: %w", ErrRuntime, e.path, err)
	}

	return file, nil
}
