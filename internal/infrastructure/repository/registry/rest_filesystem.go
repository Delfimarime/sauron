package registry

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/marketplace"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// restFactory opens a source that is a client of a registry's HTTP API.
type restFactory struct{}

// newRESTFactory builds a restFactory.
func newRESTFactory() *restFactory {
	return &restFactory{}
}

// Validate rejects options the HTTP transport does not accept: an SSH key or a
// reference. Credentials and transport security are accepted.
func (restFactory) Validate(opts ...extension.Option) error {
	options := api.Resolve(opts)

	switch {
	case options.Ref != "":
		return fmt.Errorf("%w: a reference is not supported", api.ErrUsage)
	case options.SSHKey != "":
		return fmt.Errorf("%w: an SSH key is not supported", api.ErrUsage)
	default:
		return nil
	}
}

// Open returns a read-only client of the registry's HTTP API.
func (f restFactory) Open(_ context.Context, opts ...extension.Option) (source.FileSystem, error) {
	options := api.Resolve(opts)

	if err := f.Validate(opts...); err != nil {
		return nil, err
	}

	tlsConfig, err := tlsConfigFrom(options)
	if err != nil {
		return nil, err
	}

	client, err := marketplace.New(
		marketplace.WithBaseURL(options.URI),
		marketplace.WithBasicAuth(options.Username, options.Password),
		marketplace.WithTLSConfig(tlsConfig),
		marketplace.WithTimeout(options.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", api.ErrUsage, err)
	}

	return &restFileSystem{client: client}, nil
}

// restFileSystem is a read-only client of the registry's HTTP API, backed by the
// marketplace client.
type restFileSystem struct {
	client marketplace.Client
}

// List routes uri to the matching artifact collection and returns the reported
// entries.
func (f *restFileSystem) List(ctx context.Context, uri string, opts ...source.Option) ([]source.File, error) {
	collection, err := f.collection(uri)
	if err != nil {
		return nil, err
	}

	options := source.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	list, err := collection.List(ctx, listOptionsFrom(options)...)
	if err != nil {
		return nil, convertMarketPlaceError(err)
	}

	return f.toFiles(list.Items), nil
}

// collection resolves the artifact collection a listing uri names.
func (f *restFileSystem) collection(uri string) (marketplace.ArtifactClient, error) {
	switch uri {
	case rootSkills:
		return f.client.Skills(), nil
	case rootAgents:
		return f.client.Agents(), nil
	default:
		return nil, fmt.Errorf("%w: unknown collection %q", api.ErrUsage, uri)
	}
}

// Describe is not supported by the HTTP transport yet.
func (f *restFileSystem) Describe(_ context.Context, _ string) (source.Stat, error) {
	return nil, source.ErrNotImplemented
}

// Get is not supported by the HTTP transport yet.
func (f *restFileSystem) Get(_ context.Context, _ string) (source.File, error) {
	return nil, source.ErrNotImplemented
}

// Fetch downloads the artifact archive at uri, unpacks it, and returns each
// blob as a source.File whose Name() is relative to the artifact directory.
// uri must be "<kind>/<name>", e.g. "skills/code-reviewer". The returned
// files carry the Artifact-Version header value as their Version(); an empty
// version means the registry declared none.
func (f *restFileSystem) Fetch(ctx context.Context, uri string) ([]source.File, error) {
	kind, name, err := parseArtifactURI(uri)
	if err != nil {
		return nil, err
	}

	col, err := f.collection(kind)
	if err != nil {
		return nil, err
	}

	archive, version, err := col.Content(ctx, name)
	if err != nil {
		return nil, convertMarketPlaceError(err)
	}

	return unpackArtifact(archive, uri, name, version)
}

// parseArtifactURI splits "<kind>/<name>" into its two parts, returning a
// usage error when the format is not satisfied.
func parseArtifactURI(uri string) (kind, name string, err error) {
	parts := strings.SplitN(uri, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: artifact uri must be <kind>/<name>, got %q", api.ErrUsage, uri)
	}
	return parts[0], parts[1], nil
}

// unpackArtifact gunzips and untars archive, stripping the artifact-path
// prefix from each entry and returning the blobs as source.Files carrying
// version.
func unpackArtifact(archive []byte, uri, name, version string) ([]source.File, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("%w: decompress %q: %w", api.ErrRuntime, uri, err)
	}
	defer gz.Close() //nolint:errcheck // best effort; gz is a read-only stream

	longPrefix := uri + "/"   // e.g. "skills/writer/"
	shortPrefix := name + "/" // e.g. "writer/"

	tr := tar.NewReader(gz)
	var files []source.File
	for {
		hdr, tarErr := tr.Next()
		if errors.Is(tarErr, io.EOF) {
			break
		}
		if tarErr != nil {
			return nil, fmt.Errorf("%w: read archive %q: %w", api.ErrRuntime, uri, tarErr)
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}

		relName := stripOneOf(hdr.Name, longPrefix, shortPrefix)
		if relName == "" {
			continue
		}

		content, readErr := io.ReadAll(tr)
		if readErr != nil {
			return nil, fmt.Errorf("%w: read %q from archive %q: %w", api.ErrRuntime, hdr.Name, uri, readErr)
		}

		files = append(files, restFetchedFile{
			name:    relName,
			version: version,
			size:    int64(len(content)),
			content: content,
		})
	}

	return files, nil
}

// stripOneOf trims prefix1 from s, then prefix2; returns the first stripped
// result that differs from s. Returns s unchanged when neither prefix matches.
func stripOneOf(s, prefix1, prefix2 string) string {
	if r := strings.TrimPrefix(s, prefix1); r != s {
		return r
	}
	if r := strings.TrimPrefix(s, prefix2); r != s {
		return r
	}
	return s
}

// restFetchedFile is a blob extracted from a downloaded artifact archive.
type restFetchedFile struct {
	name    string
	version string
	size    int64
	content []byte
}

// Name returns the file's path relative to the artifact directory.
func (f restFetchedFile) Name() string { return f.name }

// IsDirectory always returns false; fetched archive entries are blobs.
func (f restFetchedFile) IsDirectory() bool { return false }

// Size returns the uncompressed byte count.
func (f restFetchedFile) Size() int64 { return f.size }

// Version returns the Artifact-Version header value from the registry response.
func (f restFetchedFile) Version() string { return f.version }

// Read returns a reader over the file's content. The caller closes it.
func (f restFetchedFile) Read(_ context.Context) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.content)), nil
}

// toFiles maps artifact summaries to File entries.
func (f *restFileSystem) toFiles(items []marketplace.ArtifactSummary) []source.File {
	files := make([]source.File, 0, len(items))
	for _, it := range items {
		files = append(files, restFile{
			name:    it.Name,
			version: deref(it.Version),
			size:    derefInt64(it.Size),
		})
	}
	return files
}

// restFile is an HTTP listing entry. Its content is not readable yet.
type restFile struct {
	name    string
	version string
	size    int64
}

// Name returns the entry's name.
func (f restFile) Name() string { return f.name }

// IsDirectory reports whether the entry is a directory; HTTP entries are not.
func (f restFile) IsDirectory() bool { return false }

// Size returns the entry's size in bytes.
func (f restFile) Size() int64 { return f.size }

// Version returns the entry's version identifier.
func (f restFile) Version() string { return f.version }

// Read is not supported by the HTTP transport yet.
func (f restFile) Read(_ context.Context) (io.ReadCloser, error) {
	return nil, source.ErrNotImplemented
}
