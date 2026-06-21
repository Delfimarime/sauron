package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
)

const (
	regName = "demo"
	regURI  = "/srv/artifacts"
	envRef  = "${env:TOKEN}"
)

// TestNewAddRegistryRequestMapsFlags asserts the parsed flags and positional
// arguments land on the use case request, including the --kind default and the
// git-specific --ref.
func TestNewAddRegistryRequestMapsFlags(t *testing.T) {
	tests := []struct {
		name      string
		flags     addRegistryFlags
		args      []string
		wantKind  string
		wantRef   string
		wantUser  string
		wantTLS   bool
		wantSSH   string
		wantTmout time.Duration
	}{
		{
			name:      "defaults to http transport",
			flags:     addRegistryFlags{kindFlags: kindFlags{Kind: kindHTTP}, timeoutFlags: timeoutFlags{Timeout: 30 * time.Second}},
			args:      []string{regName, regURI},
			wantKind:  kindHTTP,
			wantTmout: 30 * time.Second,
		},
		{
			name: "git ref and credentials carried",
			flags: addRegistryFlags{
				kindFlags:     kindFlags{Kind: kindGit},
				timeoutFlags:  timeoutFlags{Timeout: 5 * time.Second},
				Ref:           "v1.2.3",
				Username:      envRef,
				SkipTLSVerify: true,
				SSHKey:        "/keys/id_ed25519",
			},
			args:      []string{regName, regURI},
			wantKind:  kindGit,
			wantRef:   "v1.2.3",
			wantUser:  envRef,
			wantTLS:   true,
			wantSSH:   "/keys/id_ed25519",
			wantTmout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var stdout bytes.Buffer

			// Act.
			request := newAddRegistryRequest(context.Background(), &tt.flags, tt.args, &stdout)

			// Assert.
			require.NotNil(t, request)
			assert.Equal(t, regName, request.Name)
			assert.Equal(t, regURI, request.URI)
			assert.Equal(t, tt.wantKind, request.Transport)
			assert.Equal(t, tt.wantRef, request.Ref)
			assert.Equal(t, tt.wantUser, request.Username)
			assert.Equal(t, tt.wantTLS, request.SkipTLSVerify)
			assert.Equal(t, tt.wantSSH, request.SSHKey)
			assert.Equal(t, tt.wantTmout, request.Timeout)
			assert.Same(t, &stdout, request.Out())
		})
	}
}

// TestAddRegistryRejectsInvalidKind asserts an unknown --kind is rejected before
// the use case runs and maps to the usage exit code.
func TestAddRegistryRejectsInvalidKind(t *testing.T) {
	// Arrange.
	flags := addRegistryFlags{kindFlags: kindFlags{Kind: "ftp"}}

	// Act.
	err := addRegistry(context.Background(), &flags, []string{regName, regURI}, &bytes.Buffer{})

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, errInvalidFlag)
	assert.Equal(t, exitUsage, exitCode(err))
}

// TestAddRegistryCommand exercises the assembled subcommand: flag binding, the
// --kind default, ExactArgs(2), and rejection of malformed input before the
// graph is built. It uses a temporary SAURON_HOME so nothing durable is touched.
func TestAddRegistryCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		wantUsage bool
	}{
		{
			name:      "rejects a missing argument",
			args:      []string{regName},
			wantErr:   true,
			wantUsage: true,
		},
		{
			name:      "rejects an unknown flag",
			args:      []string{"--nope", regName, regURI},
			wantErr:   true,
			wantUsage: true,
		},
		{
			name:      "rejects an unknown kind",
			args:      []string{"--kind", "ftp", regName, regURI},
			wantErr:   true,
			wantUsage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("SAURON_HOME", t.TempDir())

			cmd := AddRegistry()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(tt.args)
			cmd.SetContext(context.Background())

			// Act.
			err := cmd.Execute()

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantUsage {
					assert.Equal(t, exitUsage, exitCode(err))
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestAddRegistryFlagDefaults asserts every flag is registered with the
// documented default.
func TestAddRegistryFlagDefaults(t *testing.T) {
	// Arrange + Act.
	cmd := AddRegistry()

	// Assert: --kind defaults to http.
	kind, err := cmd.Flags().GetString("kind")
	require.NoError(t, err)
	assert.Equal(t, kindHTTP, kind)

	// Assert: the full flag surface is present.
	for _, name := range []string{
		"kind", "ref", "timeout", "username", "password",
		"skip-tls-verify", "ca-cert", "client-cert", "client-key", "ssh-key",
	} {
		assert.NotNilf(t, cmd.Flags().Lookup(name), "flag %q registered", name)
	}
	assert.NotNil(t, cmd.Args, "an argument validator is installed")
}

// TestExitCode covers the error-to-exit-code mapping in isolation.
func TestExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil is success", err: nil, want: exitOK},
		{name: "usage error", err: usecase.NewUsageError("x"), want: exitUsage},
		{name: "conflict error", err: usecase.NewConflictError("x"), want: exitError},
		{name: "not-found error", err: usecase.NewNotFoundError("x"), want: exitError},
		{name: "invalid flag", err: errInvalidFlag, want: exitUsage},
		{name: "generic error", err: errors.New("x"), want: exitError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, exitCode(tt.err))
		})
	}
}

// TestKindFlagsValidate covers the shared transport validator.
func TestKindFlagsValidate(t *testing.T) {
	for _, kind := range kindValues {
		f := kindFlags{Kind: kind}
		assert.NoErrorf(t, f.validate(), "kind %q is accepted", kind)
	}

	f := kindFlags{Kind: "smtp"}
	assert.ErrorIs(t, f.validate(), errInvalidFlag)
}

// TestExitCodeMapper asserts the exported mapper delegates to exitCode.
func TestExitCodeMapper(t *testing.T) {
	assert.Equal(t, exitUsage, ExitCode(usecase.NewUsageError("x")))
	assert.Equal(t, exitOK, ExitCode(nil))
}

// TestAddRegistryEndToEnd drives the assembled subcommand through the real fx
// graph against a filesystem source that hosts an artifact, asserting it
// registers and writes the confirmation to stdout. The source lives in a temp
// directory and the catalogue in a temp SAURON_HOME, so nothing durable is
// touched.
func TestAddRegistryEndToEnd(t *testing.T) {
	// Arrange: a source directory that hosts one skill artifact.
	source := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(source, ".skills", regName), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(source, ".skills", regName, "skill.yaml"),
		[]byte("placeholder\n"), 0o644,
	))
	t.Setenv("SAURON_HOME", t.TempDir())

	cmd := AddRegistry()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--kind", kindFilesystem, regName, source})

	// Act.
	err := cmd.Execute()

	// Assert.
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "registered registry \"demo\" (filesystem)")
	assert.Empty(t, stderr.String())
}

// TestAddGroup asserts the add group has no run behaviour and attaches the
// registry subcommand.
func TestAddGroup(t *testing.T) {
	// Arrange + Act.
	cmd := Add()

	// Assert.
	assert.Equal(t, "add", cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")

	var registry *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdRegistry {
			registry = sub
		}
	}
	require.NotNil(t, registry, "the registry subcommand is attached")
}
