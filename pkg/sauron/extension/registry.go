package extension

import (
	"context"
	"time"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// Registry is an artifact source Sauron can read from.
type Registry interface {
	// Validate checks that the configuration described by opts is usable
	// without opening the source.
	Validate(opts ...Option) error
	// Open establishes access to the source and returns a read-only view of
	// its contents.
	Open(ctx context.Context, opts ...Option) (source.FileSystem, error)
}

// Options configures how a Registry is validated or opened.
type Options struct {
	URI                           string
	Ref                           string
	Timeout                       time.Duration
	Username, Password            string
	SSHKey                        string
	SkipTLSVerify                 bool
	CACert, ClientCert, ClientKey string
}

// Option mutates Options.
type Option func(*Options)

// WithURI sets the source location.
func WithURI(uri string) Option {
	return func(o *Options) {
		o.URI = uri
	}
}

// WithRef sets the reference (e.g. branch, tag, or revision) to read.
func WithRef(ref string) Option {
	return func(o *Options) {
		o.Ref = ref
	}
}

// WithTimeout bounds the duration of network operations.
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithBasicAuth sets the username and password credentials.
func WithBasicAuth(user, pass string) Option {
	return func(o *Options) {
		o.Username = user
		o.Password = pass
	}
}

// WithSSHKey sets the path to the SSH private key.
func WithSSHKey(key string) Option {
	return func(o *Options) {
		o.SSHKey = key
	}
}

// WithSkipTLSVerify toggles verification of the server's TLS certificate.
func WithSkipTLSVerify(skip bool) Option {
	return func(o *Options) {
		o.SkipTLSVerify = skip
	}
}

// WithCACert sets the path to the CA certificate used to verify the server.
func WithCACert(caCert string) Option {
	return func(o *Options) {
		o.CACert = caCert
	}
}

// WithClientCert sets the paths to the client certificate and its key for
// mutual TLS.
func WithClientCert(cert, key string) Option {
	return func(o *Options) {
		o.ClientCert = cert
		o.ClientKey = key
	}
}
