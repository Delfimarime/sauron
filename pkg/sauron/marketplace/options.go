package marketplace

import (
	"crypto/tls"
	"time"
)

// config is the resolved client configuration the constructor builds resty from.
type config struct {
	baseURL   string
	username  string
	password  string
	tlsConfig *tls.Config
	timeout   time.Duration
}

// Option mutates a client config.
type Option func(*config)

// WithBaseURL sets the registry's base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *config) {
		c.baseURL = baseURL
	}
}

// WithBasicAuth sets the HTTP Basic credentials sent with every request.
func WithBasicAuth(user, pass string) Option {
	return func(c *config) {
		c.username = user
		c.password = pass
	}
}

// WithTLSConfig sets the TLS configuration used for HTTPS requests.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(c *config) {
		c.tlsConfig = tlsConfig
	}
}

// WithTimeout bounds the duration of each request.
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) {
		c.timeout = timeout
	}
}
