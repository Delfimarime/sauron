package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
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

// tlsConfigFrom assembles the TLS configuration from the options.
func tlsConfigFrom(options extension.Options) (*tls.Config, error) {
	config := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: options.SkipTLSVerify} //nolint:gosec // opt-in via SkipTLSVerify

	if options.CACert != "" {
		pool, err := caPoolFrom(options.CACert)
		if err != nil {
			return nil, err
		}
		config.RootCAs = pool
	}

	if options.ClientCert != "" || options.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(options.ClientCert, options.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("%w: load client certificate: %w", api.ErrUsage, err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	return config, nil
}

// caPoolFrom loads the certificate pool seeded with the CA at path.
func caPoolFrom(path string) (*x509.CertPool, error) {
	pem, err := os.ReadFile(path) // #nosec G304 -- certificate path is operator-supplied
	if err != nil {
		return nil, fmt.Errorf("%w: read CA certificate: %w", api.ErrUsage, err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("%w: CA certificate has no usable entries", api.ErrUsage)
	}

	return pool, nil
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
		return nil, mapError(err)
	}

	return toFiles(list.Items), nil
}

// collection resolves the artifact collection a listing uri names.
func (f *restFileSystem) collection(uri string) (marketplace.ArtifactClient, error) {
	switch strings.TrimPrefix(uri, ".") {
	case "skills":
		return f.client.Skills(), nil
	case "agents":
		return f.client.Agents(), nil
	case "personas":
		return f.client.Personas(), nil
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

// listOptionsFrom maps source listing options to marketplace list options.
func listOptionsFrom(options source.Options) []marketplace.ListOption {
	var opts []marketplace.ListOption
	if options.Search != nil {
		opts = append(opts, marketplace.WithSearch(*options.Search))
	}
	if options.Sort != nil {
		opts = append(opts, marketplace.WithSort(*options.Sort))
	}
	if options.Limit != nil {
		opts = append(opts, marketplace.WithLimit(*options.Limit))
	}
	if options.Offset != nil {
		opts = append(opts, marketplace.WithOffset(*options.Offset))
	}
	return opts
}

// mapError classifies a marketplace error as a usage or runtime failure.
func mapError(err error) error {
	switch {
	case marketplace.IsUnauthorized(err), marketplace.IsForbidden(err), errors.Is(err, marketplace.ErrInvalidConfig):
		return fmt.Errorf("%w: %w", api.ErrUsage, err)
	default:
		return fmt.Errorf("%w: %w", api.ErrRuntime, err)
	}
}

// toFiles maps artifact summaries to File entries.
func toFiles(items []marketplace.ArtifactSummary) []source.File {
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

// deref returns the pointed-to string, or the empty string when nil.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefInt64 returns the pointed-to value, or zero when nil.
func derefInt64(n *int64) int64 {
	if n == nil {
		return 0
	}
	return *n
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
