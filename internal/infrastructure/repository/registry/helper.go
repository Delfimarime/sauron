package registry

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/marketplace"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

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

// listOptionsFrom maps source listing options to marketplace list options.
func listOptionsFrom(options source.Options) []marketplace.ListOption {
	var opts []marketplace.ListOption
	if options.Search != nil {
		opts = append(opts, marketplace.WithSearch(*options.Search))
	}
	if options.Sort != nil {
		opts = append(opts, marketplace.WithSort(signedSort(*options.Sort, options.Order)))
	}
	if options.Limit != nil {
		opts = append(opts, marketplace.WithLimit(*options.Limit))
	}
	if options.Offset != nil {
		opts = append(opts, marketplace.WithOffset(*options.Offset))
	}
	return opts
}

// signedSort renders the registry HTTP API's signed sort directive from a sort
// field and direction: "-name" when order is "desc", "+name" otherwise.
func signedSort(sortField string, order *string) string {
	if order != nil && *order == "desc" {
		return "-" + sortField
	}
	return "+" + sortField
}

// convertMarketPlaceError classifies a marketplace error as a usage or runtime failure.
func convertMarketPlaceError(err error) error {
	switch {
	case marketplace.IsUnauthorized(err), marketplace.IsForbidden(err), errors.Is(err, marketplace.ErrInvalidConfig):
		return fmt.Errorf("%w: %w", api.ErrUsage, err)
	default:
		return fmt.Errorf("%w: %w", api.ErrRuntime, err)
	}
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
