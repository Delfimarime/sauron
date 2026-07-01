package registry

import (
	"context"
	"fmt"
	"io"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// gitTreeSource is a read-only source.FileSystem backed by the tree objects of a
// cloned git repository held in memory. An artifact's native version is the git
// tree-object hash of its directory.
type gitTreeSource struct {
	repo *gogit.Repository
	root *object.Tree
}

var _ source.FileSystem = (*gitTreeSource)(nil)

// newGitTreeSource builds a gitTreeSource over the commit tree root, retaining
// repo so its in-memory object store stays reachable.
func newGitTreeSource(repo *gogit.Repository, root *object.Tree) *gitTreeSource {
	return &gitTreeSource{repo: repo, root: root}
}

// List returns the immediate directory children of the tree at uri, applying any
// paging given via opts through the shared helper. An artifact is a directory
// under skills/ or agents/, so a stray blob sibling is excluded — keeping a file
// from becoming a catalogue or install entry carrying a blob hash as its version.
func (s *gitTreeSource) List(_ context.Context, uri string, opts ...source.Option) ([]source.File, error) {
	tree, err := s.subtree(uri)
	if err != nil {
		return nil, err
	}

	options := source.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	entries := make([]source.File, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		if entry.Mode != filemode.Dir {
			continue
		}
		entries = append(entries, gitEntry{
			name:    entry.Name,
			isDir:   true,
			version: entry.Hash.String(),
		})
	}

	return api.Page(entries, options), nil
}

// Fetch walks the subtree at uri recursively and returns each blob as a File
// whose Name is its path relative to that subtree. Every returned file carries
// the subtree's tree-object hash as its version.
func (s *gitTreeSource) Fetch(_ context.Context, uri string) ([]source.File, error) {
	tree, err := s.subtree(uri)
	if err != nil {
		return nil, err
	}

	version := tree.Hash.String()
	files := make([]source.File, 0)
	walkErr := tree.Files().ForEach(func(f *object.File) error {
		blob := f.Blob
		files = append(files, gitEntry{
			name:    f.Name,
			size:    f.Size,
			version: version,
			blob:    &blob,
		})
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("%w: fetch %q: %w", api.ErrRuntime, uri, walkErr)
	}

	return files, nil
}

// Describe is not supported by the git transport yet.
func (s *gitTreeSource) Describe(_ context.Context, _ string) (source.Stat, error) {
	return nil, source.ErrNotImplemented
}

// Get is not supported by the git transport yet.
func (s *gitTreeSource) Get(_ context.Context, _ string) (source.File, error) {
	return nil, source.ErrNotImplemented
}

// subtree resolves the tree at uri; an empty uri resolves to the commit root.
func (s *gitTreeSource) subtree(uri string) (*object.Tree, error) {
	if uri == "" || uri == "." {
		return s.root, nil
	}

	tree, err := s.root.Tree(uri)
	if err != nil {
		return nil, fmt.Errorf("%w: read %q: %w", api.ErrRuntime, uri, err)
	}

	return tree, nil
}

// gitEntry is a git tree entry exposing its metadata and, for blobs, content.
type gitEntry struct {
	name    string
	isDir   bool
	size    int64
	version string
	blob    *object.Blob
}

// Name returns the entry's name.
func (e gitEntry) Name() string { return e.name }

// IsDirectory reports whether the entry is a directory.
func (e gitEntry) IsDirectory() bool { return e.isDir }

// Size returns the entry's size in bytes.
func (e gitEntry) Size() int64 { return e.size }

// Version returns the entry's tree-object hash.
func (e gitEntry) Version() string { return e.version }

// Read opens the entry's blob content; the caller closes the returned reader.
// Directories are not readable.
func (e gitEntry) Read(_ context.Context) (io.ReadCloser, error) {
	if e.isDir {
		return nil, fmt.Errorf("%w: read %q: is a directory", api.ErrRuntime, e.name)
	}

	blob := e.blob
	if blob == nil {
		return nil, fmt.Errorf("%w: read %q: no content", api.ErrRuntime, e.name)
	}

	reader, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("%w: read %q: %w", api.ErrRuntime, e.name, err)
	}

	return reader, nil
}
