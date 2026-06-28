//go:build unit

package host

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

func TestHostRuntimeIsReadOnly(t *testing.T) {
	assert.True(t, New("/bin/echo", t.TempDir()).IsReadOnly())
}

func TestHostRuntimeStartCreatesHome(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "home")
	h := New("/bin/echo", dir)
	require.NoError(t, h.Start(context.Background()))
	require.NoError(t, h.Stop(context.Background()))

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestHostRuntimeExecuteSuccess(t *testing.T) {
	code, out, err := New("/bin/echo", t.TempDir()).Execute(context.Background(), "hello")
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "hello\n", out)
}

func TestHostRuntimeExecuteNonZeroReturnsStderr(t *testing.T) {
	// /bin/sh is present on macOS and Linux; -c runs the script as args.
	code, out, err := New("/bin/sh", t.TempDir()).
		Execute(context.Background(), "-c", "echo oops 1>&2; exit 2")
	require.NoError(t, err) // a non-zero exit is NOT a harness error
	assert.Equal(t, 2, code)
	assert.Equal(t, "oops\n", out)
}

func TestHostRuntimeExecutePinsSauronHome(t *testing.T) {
	dir := t.TempDir()
	code, out, err := New("/bin/sh", dir).
		Execute(context.Background(), "-c", "printf %s \"$SAURON_HOME\"")
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, dir, out, "the binary runs with $SAURON_HOME pinned to the per-scenario dir")
}

func TestHostRuntimeExecuteMissingBinaryErrors(t *testing.T) {
	_, _, err := New("/no/such/binary", t.TempDir()).Execute(context.Background())
	assert.Error(t, err)
}

func TestHostRuntimeCopyToReadFileRoundTrip(t *testing.T) {
	ctx := context.Background()
	h := New("/bin/echo", t.TempDir())

	// A home-relative path round-trips under the per-scenario directory.
	require.NoError(t, h.CopyTo(ctx, "registries.yaml", []byte("kind: Registry")))
	got, err := h.ReadFile(ctx, "registries.yaml")
	require.NoError(t, err)
	assert.Equal(t, []byte("kind: Registry"), got)

	// A "~/" path is anchored at the same per-scenario home, not the real home.
	require.NoError(t, h.CopyTo(ctx, "~/nested/settings.yaml", []byte("kind: Provider")))
	got, err = h.ReadFile(ctx, "~/nested/settings.yaml")
	require.NoError(t, err)
	assert.Equal(t, []byte("kind: Provider"), got)
}

func TestHostRuntimeReadFileMissingErrors(t *testing.T) {
	_, err := New("/bin/echo", t.TempDir()).ReadFile(context.Background(), "absent.yaml")
	assert.Error(t, err)
}

func TestHostRuntimeFolderMaterializesContent(t *testing.T) {
	ctx := context.Background()
	h := New("/bin/echo", t.TempDir())

	folder := h.Folder("default")
	folder.Expose(runtime.Resource{Path: ".skills/go-style/skill.yaml", Content: []byte("name: go-style")})

	path, err := folder.Path(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(path, ".skills/go-style/skill.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "name: go-style", string(data))
}

func TestHostRuntimeFolderIsStablePerAlias(t *testing.T) {
	h := New("/bin/echo", t.TempDir())
	h.Folder("default").Expose(runtime.Resource{Path: "a", Content: []byte("1")})
	h.Folder("default").Expose(runtime.Resource{Path: "b", Content: []byte("2")})

	path, err := h.Folder("default").Path(context.Background())
	require.NoError(t, err)
	for _, name := range []string{"a", "b"} {
		_, err := os.Stat(filepath.Join(path, name))
		require.NoError(t, err, "both Expose calls accumulate on the same source")
	}
}

// TestHostRuntimeWebserverServesInProcess verifies the host runtime serves an http
// registry source in-process at 127.0.0.1; the same Webserver source accumulates
// across calls and is torn down by Stop.
func TestHostRuntimeWebserverServesInProcess(t *testing.T) {
	ctx := context.Background()
	h := New("/bin/echo", t.TempDir())
	t.Cleanup(func() { _ = h.Stop(ctx) })

	url, err := h.Webserver("default").URL(ctx)
	require.NoError(t, err)
	assert.Regexp(t, `^http://127\.0\.0\.1:\d+$`, url, "the host binary reaches the fixture on loopback")

	again, err := h.Webserver("default").URL(ctx)
	require.NoError(t, err)
	assert.Equal(t, url, again, "the per-alias source is created once and reused")
}

func TestHostRuntimeGitErrors(t *testing.T) {
	ctx := context.Background()
	h := New("/bin/echo", t.TempDir())

	_, err := h.Git("default").URL(ctx)
	assert.Error(t, err, "git is not available on the host runtime")
}
