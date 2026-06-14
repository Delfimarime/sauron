//go:build unit

package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestNewPrependsMainService(t *testing.T) {
	rt, err := New("/host/sauron", "/tmp/work")
	require.NoError(t, err)

	require.Len(t, rt.specs, 1)
	main := rt.specs[0]
	assert.Equal(t, mainService, main.Service)
	assert.Equal(t, "alpine:3", main.Image)
	assert.Equal(t, "tail -f /dev/null", main.Entrypoint)
	assert.Equal(t, "/root/.sauron/", main.Env["SAURON_HOME"])
	require.Len(t, main.Mount, 1)
	assert.Equal(t, FileSpec{SourceFile: "/host/sauron", Path: sauronPath}, main.Mount[0])
	assert.False(t, rt.IsReadOnly())
}

func TestNewAppendsDependenciesAndExtraFiles(t *testing.T) {
	rt, err := New("/host/sauron", "/tmp/work",
		WithContainer(ContainerSpec{Service: "git", Image: "gitea/gitea:1"}),
		WithFile(FileSpec{SourceFile: "/host/known_hosts", Path: "/root/.ssh/known_hosts"}),
	)
	require.NoError(t, err)

	require.Len(t, rt.specs, 2)
	assert.Equal(t, mainService, rt.specs[0].Service)
	assert.Equal(t, "git", rt.specs[1].Service)

	// WithFile mounts land on main, ahead of the binary mount New appends.
	require.Len(t, rt.specs[0].Mount, 2)
	assert.Equal(t, "/host/known_hosts", rt.specs[0].Mount[0].SourceFile)
	assert.Equal(t, sauronPath, rt.specs[0].Mount[1].Path)
}

func TestNewRejectsReservedMainService(t *testing.T) {
	_, err := New("/host/sauron", "/tmp/work",
		WithContainer(ContainerSpec{Service: mainService, Image: "x"}))
	assert.Error(t, err)
}

func TestExecuteRejectsEmptyArgs(t *testing.T) {
	rt, err := New("/host/sauron", "/tmp/work")
	require.NoError(t, err)
	_, _, err = rt.Execute(context.Background())
	assert.Error(t, err)
}

// TestDockerRuntimeExecutesMountedBinary drives the full lifecycle against a real
// Docker daemon: it mounts a fake "sauron" into main, starts the stack, and checks
// both the arg0 rewrite (stdout) and the stderr/exit-code path. Skipped when no
// Docker provider is available.
func TestDockerRuntimeExecutesMountedBinary(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()

	bin := filepath.Join(t.TempDir(), "sauron")
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\necho \"sauron v9.9.9\"\n"), 0o755))

	rt, err := New(bin, t.TempDir())
	require.NoError(t, err)
	require.NoError(t, rt.Start(ctx))
	t.Cleanup(func() { _ = rt.Stop(ctx) })

	// arg0 "sauron" is rewritten to the mounted path and run inside main.
	code, out, err := rt.Execute(ctx, "sauron", "--version")
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "sauron v9.9.9\n", out)

	// A non-sauron command: stderr is returned on a non-zero exit.
	code, out, err = rt.Execute(ctx, "sh", "-c", "echo boom 1>&2; exit 4")
	require.NoError(t, err)
	assert.Equal(t, 4, code)
	assert.Equal(t, "boom\n", out)
}
