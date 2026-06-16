package http

import (
	"net/http"
)

type BasicAuthRoundTrip struct {
	username string
	password string
	base     http.RoundTripper
}

func NewBasicAuthRoundTrip(rt http.RoundTripper, username, password string) *BasicAuthRoundTrip {
	return &BasicAuthRoundTrip{
		base:     rt,
		username: username,
		password: password,
	}
}

func (rt *BasicAuthRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	rq := req.Clone(req.Context())
	if rq.Header.Get("Authorization") == "" {
		rq.SetBasicAuth(rt.username, rt.password)
	}
	return rt.base.RoundTrip(rq)
}

func WithBasicAuth(username, password string) func(*http.Client) error {
	return func(c *http.Client) error {
		c.Transport = NewBasicAuthRoundTrip(c.Transport, username, password)
		return nil
	}
}
