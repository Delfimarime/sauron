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

func WithTLS(serverName, trustStoreURI string, skipVerify bool) (func(*http.Transport), error) {
	var roots *x509.CertPool
	sys, err := x509.SystemCertPool()
	if err == nil && sys != nil {
		roots = sys
	} else {
		roots = x509.NewCertPool()
	}
	if trustStoreURI != "" {
		pem, prob := os.ReadFile(trustStoreURI)
		if prob != nil {
			if pe, ok := errors.AsType[*fs.PathError](prob); ok {
				return nil, fmt.Errorf("cannot read truststore %q: %w", trustStoreURI, pe)
			}
			return nil, prob
		}
		if ok := roots.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("invalid PEM contents in truststore %q", trustStoreURI)
		}
	}
	return func(t *http.Transport) {
		t.TLSClientConfig = &tls.Config{
			RootCAs:            roots,
			InsecureSkipVerify: skipVerify,
			ServerName:         serverName,
			MinVersion:         tls.VersionTLS12,
		}
	}, nil
}

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
