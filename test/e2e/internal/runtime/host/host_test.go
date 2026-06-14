package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostRuntimeIsReadOnly(t *testing.T) {
	assert.True(t, New("/bin/echo").IsReadOnly())
}

func TestHostRuntimeStartStopAreNoOps(t *testing.T) {
	h := New("/bin/echo")
	require.NoError(t, h.Start(context.Background()))
	require.NoError(t, h.Stop(context.Background()))
}

func TestHostRuntimeExecuteSuccess(t *testing.T) {
	code, out, err := New("/bin/echo").Execute(context.Background(), "hello")
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "hello\n", out)
}

func TestHostRuntimeExecuteNonZeroReturnsStderr(t *testing.T) {
	// /bin/sh is present on macOS and Linux; -c runs the script as args.
	code, out, err := New("/bin/sh").
		Execute(context.Background(), "-c", "echo oops 1>&2; exit 2")
	require.NoError(t, err) // a non-zero exit is NOT a harness error
	assert.Equal(t, 2, code)
	assert.Equal(t, "oops\n", out)
}

func TestHostRuntimeExecuteMissingBinaryErrors(t *testing.T) {
	_, _, err := New("/no/such/binary").Execute(context.Background())
	assert.Error(t, err)
}
