package http

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"
)

// errInvalidTrustStore reports a truststore file whose contents are not valid PEM.
var errInvalidTrustStore = errors.New("invalid PEM contents in truststore")

// WithSimpleRoundTripper installs an *http.Transport on the client, applying each
// option to it.
func WithSimpleRoundTripper(opts ...func(*http.Transport) error) func(*http.Client) error {
	return func(c *http.Client) error {
		transport := &http.Transport{
			ForceAttemptHTTP2: true,
			Proxy:             http.ProxyFromEnvironment,
		}
		for _, f := range opts {
			if err := f(transport); err != nil {
				return err
			}
		}
		c.Transport = transport
		return nil
	}
}

// NewSimpleRoundTripper builds an *http.Transport configured by the given options.
func NewSimpleRoundTripper(
	opts ...func(*http.Transport),
) (http.RoundTripper, error) {
	transport := &http.Transport{
		ForceAttemptHTTP2: true,
		Proxy:             http.ProxyFromEnvironment,
	}
	for _, f := range opts {
		f(transport)
	}
	return transport, nil
}

// WithTLS configures the transport's TLS: optional trust roots loaded from
// trustStoreURI, the expected server name, and an opt-in skip-verify.
func WithTLS(serverName, trustStoreURI string, skipVerify bool) (func(*http.Transport), error) {
	var roots *x509.CertPool
	sys, err := x509.SystemCertPool()
	if err == nil && sys != nil {
		roots = sys
	} else {
		roots = x509.NewCertPool()
	}
	if trustStoreURI != "" {
		pem, prob := os.ReadFile(trustStoreURI) //nolint:gosec // trustStoreURI is an operator-provided configuration path, not attacker-controlled input
		if prob != nil {
			if pe, ok := errors.AsType[*fs.PathError](prob); ok {
				return nil, fmt.Errorf("cannot read truststore %q: %w", trustStoreURI, pe)
			}
			return nil, prob
		}
		if ok := roots.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("%w: %q", errInvalidTrustStore, trustStoreURI)
		}
	}
	return func(t *http.Transport) {
		t.TLSClientConfig = &tls.Config{
			RootCAs:            roots,
			InsecureSkipVerify: skipVerify, //nolint:gosec // skipVerify is an explicit operator opt-in for non-production trust
			ServerName:         serverName,
			MinVersion:         tls.VersionTLS12,
		}
	}, nil
}

// WithConnectionPool configures the transport's idle-connection limits and idle timeout.
func WithConnectionPool(
	maximumNumberOfIdleConnections int,
	maximumNumberOfConnectionsPerHost int,
	idleConnectionTimeout time.Duration,
) (func(*http.Transport), error) {
	return func(transport *http.Transport) {
		if idleConnectionTimeout > 0 {
			transport.IdleConnTimeout = idleConnectionTimeout
		}
		if maximumNumberOfIdleConnections > 0 {
			transport.MaxIdleConns = maximumNumberOfIdleConnections
		}
		if maximumNumberOfConnectionsPerHost > 0 {
			transport.MaxIdleConnsPerHost = maximumNumberOfConnectionsPerHost
		}
	}, nil
}

// WithTimeouts configures the transport's TLS-handshake, response-header, and
// expect-continue timeouts.
func WithTimeouts(
	handshakeTimeout time.Duration,
	responseHeaderTimeout time.Duration,
	expectContinueTimeout time.Duration,
) (func(*http.Transport), error) {
	return func(transport *http.Transport) {
		if handshakeTimeout > 0 {
			transport.TLSHandshakeTimeout = handshakeTimeout
		}
		if responseHeaderTimeout > 0 {
			transport.ResponseHeaderTimeout = responseHeaderTimeout
		}
		if expectContinueTimeout > 0 {
			transport.ExpectContinueTimeout = expectContinueTimeout
		}
	}, nil
}
