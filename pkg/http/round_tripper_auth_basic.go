package http

import (
	"net/http"
)

// BasicAuthRoundTrip is an http.RoundTripper that adds HTTP Basic credentials to
// each request that does not already carry an Authorization header.
type BasicAuthRoundTrip struct {
	username string
	password string
	base     http.RoundTripper
}

// NewBasicAuthRoundTrip wraps rt with one that sets Basic auth credentials.
func NewBasicAuthRoundTrip(rt http.RoundTripper, username, password string) *BasicAuthRoundTrip {
	return &BasicAuthRoundTrip{
		base:     rt,
		username: username,
		password: password,
	}
}

// RoundTrip adds Basic auth credentials, unless the request already has an
// Authorization header, and delegates to the base round-tripper.
func (rt *BasicAuthRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	rq := req.Clone(req.Context())
	if rq.Header.Get("Authorization") == "" {
		rq.SetBasicAuth(rt.username, rt.password)
	}
	return rt.base.RoundTrip(rq)
}

// WithBasicAuth installs a Basic-auth round-tripper on the client.
func WithBasicAuth(username, password string) func(*http.Client) error {
	return func(c *http.Client) error {
		c.Transport = NewBasicAuthRoundTrip(c.Transport, username, password)
		return nil
	}
}
