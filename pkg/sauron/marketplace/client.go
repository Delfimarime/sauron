package marketplace

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Client is a fluent entry point to a registry's artifact collections.
type Client interface {
	// Skills returns the client for the registry's skill artifacts.
	Skills() ArtifactClient
	// Agents returns the client for the registry's agent artifacts.
	Agents() ArtifactClient
	// Personas returns the client for the registry's persona artifacts.
	Personas() ArtifactClient
}

// client is the resty-backed Client implementation.
type client struct {
	rest *resty.Client
}

// New builds a Client from the supplied options.
func New(opts ...Option) (Client, error) {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	rest, err := newResty(cfg)
	if err != nil {
		return nil, err
	}

	return &client{rest: rest}, nil
}

// newResty configures the underlying resty client from cfg.
func newResty(cfg *config) (*resty.Client, error) {
	if _, err := url.Parse(cfg.baseURL); err != nil {
		return nil, fmt.Errorf("%w: parse base url %q: %w", ErrInvalidConfig, cfg.baseURL, err)
	}

	rest := resty.New().
		SetBaseURL(strings.TrimRight(cfg.baseURL, "/")).
		SetTimeout(cfg.timeout)

	if cfg.tlsConfig != nil {
		rest.SetTLSClientConfig(cfg.tlsConfig)
	}
	if cfg.username != "" || cfg.password != "" {
		rest.SetBasicAuth(cfg.username, cfg.password)
	}

	return rest, nil
}

// Skills returns the client for skill artifacts.
func (c *client) Skills() ArtifactClient {
	return &artifactClient{rest: c.rest, kind: kindSkills}
}

// Agents returns the client for agent artifacts.
func (c *client) Agents() ArtifactClient {
	return &artifactClient{rest: c.rest, kind: kindAgents}
}

// Personas returns the client for persona artifacts.
func (c *client) Personas() ArtifactClient {
	return &artifactClient{rest: c.rest, kind: kindPersonas}
}
