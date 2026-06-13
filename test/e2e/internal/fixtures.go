package internal

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// Container is the harness's handle on a Testcontainers-provisioned dependency.
// Scenarios that need a hermetic git or HTTP endpoint provision it through this
// type rather than reaching the public internet, keeping the gate
// self-contained. The smoke scenario needs no container; richer fixtures land
// with feature coverage once the ssh/fixture strategy (ADR-0002 remotes are
// ssh-only) is settled.
type Container struct {
	ctr testcontainers.Container
}

// Start provisions a container from req and returns a handle bound to it. It is
// the single entrypoint scenarios use to obtain an ephemeral dependency, so the
// testcontainers API stays behind one harness seam.
func Start(ctx context.Context, req testcontainers.ContainerRequest) (*Container, error) {
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &Container{ctr: ctr}, nil
}

// Terminate stops and removes the underlying container.
func (c *Container) Terminate() error {
	return testcontainers.TerminateContainer(c.ctr)
}
