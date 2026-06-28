package httpregistry

import (
	"context"
	"fmt"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// Source adapts a Server to the runtime.Source contract. Expose accumulates onto the
// server; URL lazily starts it and returns the runtime-appropriate address (the
// caller passes 127.0.0.1 on the host runtime, host.docker.internal on the docker
// runtime). The remaining accessors are capability gaps for an http source and return
// errors, honouring the harness's gap-as-error principle.
type Source struct {
	server *Server
	host   string
}

// NewSource wraps server, advertising it at host.
func NewSource(server *Server, host string) *Source {
	return &Source{server: server, host: host}
}

// Server returns the underlying server so a runtime can read its auth binding and
// drive its lifecycle.
func (s *Source) Server() *Server { return s.server }

// Expose forwards accumulated content and auth to the server.
func (s *Source) Expose(resources ...runtime.Resource) { s.server.Expose(resources...) }

// URL starts the server (idempotently) and returns its advertised address.
func (s *Source) URL(_ context.Context) (string, error) {
	if err := s.server.Start(); err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%d", s.host, s.server.Port()), nil
}

// Path is not meaningful for an http source.
func (s *Source) Path(context.Context) (string, error) {
	return "", fmt.Errorf("http registry source has no path; use its url")
}

// SSHKey is not meaningful for an http source.
func (s *Source) SSHKey(context.Context) (string, error) {
	return "", fmt.Errorf("http registry source has no ssh key; use its url")
}

// Revision is not meaningful for an http source.
func (s *Source) Revision(context.Context) (string, error) {
	return "", fmt.Errorf("http registry source has no revision; use its url")
}
