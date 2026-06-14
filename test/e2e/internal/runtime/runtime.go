package runtime

import "context"

// Runtime is where the command-under-test executes: the host OS, or a sandbox.
// It owns execution, so in the container/compose cases the binary runs inside the
// sandbox and reaches dependencies by service name on the internal network — which
// is why no Endpoint/Env plumbing crosses the boundary.
type Runtime interface {
	// IsReadOnly reports whether the runtime must not be mutated by a scenario.
	// The host OS is read-only (true); ephemeral sandboxes are not (false).
	IsReadOnly() bool
	// Execute runs the binary under test with command as its args, returning the
	// exit code, the relevant output stream (stdout on success, stderr on a
	// non-zero exit), and an error ONLY for harness-level failures (the process or
	// container exec could not run). A non-zero exit is not an error.
	Execute(ctx context.Context, command ...string) (int, string, error)

	Start(ctx context.Context) error

	Stop(ctx context.Context) error // tear everything down
}
