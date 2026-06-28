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

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/httpregistry"
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

func TestDockerGitSourceYieldsSSHURL(t *testing.T) {
	rt, err := New("/host/sauron", "/tmp/work")
	require.NoError(t, err)
	url, err := rt.Git("default").URL(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "ssh://git@registry-git-default:22/home/git/acme.git", url)
}

func TestContainerPath(t *testing.T) {
	assert.Equal(t, "/root/.sauron/registries.yaml", containerPath("registries.yaml"))
	assert.Equal(t, "/root/x.yaml", containerPath("~/x.yaml"))
	assert.Equal(t, "/etc/passwd", containerPath("/etc/passwd"))
}

// TestBuildSpecsFoldsSources verifies the pure, pre-Up accumulation: a folder source
// adds content mounts onto "main", and a webserver source (served in-process, not as a
// sidecar) wires the host gateway and binds its ${env:VAR} basic-auth secret onto
// "main".
func TestBuildSpecsFoldsSources(t *testing.T) {
	rt, err := New("/host/sauron", "/tmp/work")
	require.NoError(t, err)

	rt.Folder("default").Expose(runtime.Resource{Path: "skills/go/skill.yaml", Content: []byte("name: go")})
	rt.Webserver("default").Expose(runtime.Resource{Path: "skills/go/skill.yaml", Content: []byte("name: go")})
	rt.Webserver("default").Expose(runtime.Resource{Username: "acme", Password: "${env:ACME_TOKEN}"})

	specs, err := buildSpecs(rt.specs, rt.folders, rt.webservers, rt.gits)
	require.NoError(t, err)

	main := findService(t, specs, mainService)
	assert.True(t, hasMountAt(main, folderPath("default")+"/skills/go/skill.yaml"),
		"folder content is mounted into main")
	assert.Contains(t, main.ExtraHosts, hostGatewayExtraHost,
		"the host gateway lets the container reach the in-process fixture")
	assert.Equal(t, httpregistry.Secret, main.Env["ACME_TOKEN"],
		"the ${env:VAR} basic-auth secret is bound on the binary's container")

	// No webserver sidecar: the fixture runs in the test process, so specs is just main.
	assert.Len(t, specs, 1, "a webserver source adds no container")

	// buildSpecs is pure: the runtime's own specs are not mutated.
	assert.Equal(t, mainService, rt.specs[0].Service)
	assert.Len(t, rt.specs, 1, "buildSpecs does not append onto rt.specs")
	assert.Empty(t, rt.specs[0].ExtraHosts, "the input base's main is not mutated")
}

// TestBuildSpecsFoldsGitSource verifies a git source adds an sshd sidecar serving a
// seeded bare repo and mounts the matching client key material into "main".
func TestBuildSpecsFoldsGitSource(t *testing.T) {
	rt, err := New("/host/sauron", "/tmp/work")
	require.NoError(t, err)

	rt.Git("default").Expose(runtime.Resource{Path: "skills/go/skill.yaml", Content: []byte("name: go")})

	specs, err := buildSpecs(rt.specs, rt.folders, rt.webservers, rt.gits)
	require.NoError(t, err)

	srv := findService(t, specs, gitService("default"))
	assert.Equal(t, gitImage, srv.Image)
	assert.True(t, hasMountAt(srv, gitSeedDir+"/skills/go/skill.yaml"), "seed content is mounted")
	assert.True(t, hasMountAt(srv, gitAuthKeys), "client public key is installed as authorized_keys")
	assert.True(t, hasMountAt(srv, gitHostKey), "the server host key is pinned")
	assert.True(t, hasMountAt(srv, gitEntrypoint), "the seed entrypoint is mounted")

	main := findService(t, specs, mainService)
	assert.True(t, hasMountAt(main, gitClientKey), "the private key is mounted into main")
	assert.True(t, hasMountAt(main, gitKnownHosts), "the pinned host key is a known_hosts entry")
	assert.True(t, hasMountAt(main, gitSSHConfPath), "an ssh config selects the key for the sidecar")
}

func findService(t *testing.T, specs []ContainerSpec, name string) ContainerSpec {
	t.Helper()
	for _, s := range specs {
		if s.Service == name {
			return s
		}
	}
	require.Failf(t, "service not found", "no %q in specs", name)
	return ContainerSpec{}
}

func hasMountAt(spec ContainerSpec, path string) bool {
	for _, m := range spec.Mount {
		if m.Path == path {
			return true
		}
	}
	return false
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
