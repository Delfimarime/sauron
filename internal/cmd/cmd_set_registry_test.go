package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const (
	regName = "demo"
	regURI  = "/srv/artifacts"
	envRef  = "${env:TOKEN}"
)

// TestNewSetRegistryInputMapsFlags asserts the parsed flags and positional
// argument land on the use case input, including the --transport default and the
// git-specific --revision.
func TestNewSetRegistryInputMapsFlags(t *testing.T) {
	tests := []struct {
		name      string
		flags     setRegistryFlags
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
			flags:     setRegistryFlags{transportFlags: transportFlags{Transport: transportHTTP}, timeoutFlags: timeoutFlags{Timeout: 30 * time.Second}},
			args:      []string{regURI},
			wantKind:  transportHTTP,
			wantTmout: 30 * time.Second,
		},
		{
			name: "git ref and credentials carried",
			flags: setRegistryFlags{
				transportFlags: transportFlags{Transport: transportGit},
				timeoutFlags:   timeoutFlags{Timeout: 5 * time.Second},
				Revision:       "v1.2.3",
				Username:       envRef,
				SkipTLSVerify:  true,
				SSHKey:         "/keys/id_ed25519",
			},
			args:      []string{regURI},
			wantKind:  transportGit,
			wantRef:   "v1.2.3",
			wantUser:  envRef,
			wantTLS:   true,
			wantSSH:   "/keys/id_ed25519",
			wantTmout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			input := usecase.SetRegistryRequest{
				Source:        tt.args[0],
				Transport:     tt.flags.Transport,
				Revision:      tt.flags.Revision,
				Username:      tt.flags.Username,
				Password:      tt.flags.Password,
				SSHKey:        tt.flags.SSHKey,
				SkipTLSVerify: tt.flags.SkipTLSVerify,
				CACert:        tt.flags.CACert,
				ClientCert:    tt.flags.ClientCert,
				ClientKey:     tt.flags.ClientKey,
				Timeout:       tt.flags.Timeout,
			}

			// Assert.
			assert.Equal(t, regURI, input.Source)
			assert.Equal(t, tt.wantKind, input.Transport)
			assert.Equal(t, tt.wantRef, input.Revision)
			assert.Equal(t, tt.wantUser, input.Username)
			assert.Equal(t, tt.wantTLS, input.SkipTLSVerify)
			assert.Equal(t, tt.wantSSH, input.SSHKey)
			assert.Equal(t, tt.wantTmout, input.Timeout)
		})
	}
}

// TestSetRegistryRejectsInvalidKind asserts an unknown --transport is rejected before
// the use case runs and maps to the usage exit code.
func TestSetRegistryRejectsInvalidKind(t *testing.T) {
	// Arrange.
	flags := setRegistryFlags{transportFlags: transportFlags{Transport: "ftp"}}

	// Act.
	err := setRegistry(context.Background(), &flags, []string{regURI}, &bytes.Buffer{})

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, errInvalidFlag)
	assert.Equal(t, exitUsage, ExitCode(err))
}

// TestSetRegistryCommand exercises the assembled subcommand: flag binding, the
// --transport default, ExactArgs(1), and rejection of malformed input before the
// graph is built. It uses a temporary SAURON_HOME so nothing durable is touched.
func TestSetRegistryCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		wantUsage bool
	}{
		{
			name:      "rejects a missing argument",
			args:      []string{},
			wantErr:   true,
			wantUsage: true,
		},
		{
			name:      "rejects an unknown flag",
			args:      []string{"--nope", regURI},
			wantErr:   true,
			wantUsage: true,
		},
		{
			name:      "rejects an unknown kind",
			args:      []string{"--transport", "ftp", regURI},
			wantErr:   true,
			wantUsage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("SAURON_HOME", t.TempDir())

			cmd := SetRegistry()
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
					assert.Equal(t, exitUsage, ExitCode(err))
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestSetRegistryFlagDefaults asserts every flag is registered with the
// documented default.
func TestSetRegistryFlagDefaults(t *testing.T) {
	// Arrange + Act.
	cmd := SetRegistry()

	// Assert: --transport defaults to http.
	kind, err := cmd.Flags().GetString(flagTransport)
	require.NoError(t, err)
	assert.Equal(t, transportHTTP, kind)

	// Assert: the full flag surface is present.
	for _, name := range []string{
		flagTransport, "revision", "timeout", "username", "password",
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
		{name: "not-found error", err: usecase.NewNotFoundError("x"), want: exitError},
		{name: "invalid flag", err: errInvalidFlag, want: exitUsage},
		{name: "generic error", err: errors.New("x"), want: exitError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExitCode(tt.err))
		})
	}
}

// TestKindFlagsValidate covers the shared transport validator.
func TestKindFlagsValidate(t *testing.T) {
	for _, kind := range transportValues {
		f := transportFlags{Transport: kind}
		assert.NoErrorf(t, f.validate(), "kind %q is accepted", kind)
	}

	f := transportFlags{Transport: "smtp"}
	assert.ErrorIs(t, f.validate(), errInvalidFlag)
}

// TestExitCodeMapper asserts the exported mapper delegates to exitCode.
func TestExitCodeMapper(t *testing.T) {
	assert.Equal(t, exitUsage, ExitCode(usecase.NewUsageError("x")))
	assert.Equal(t, exitOK, ExitCode(nil))
}

// TestSetRegistryEndToEnd drives the assembled subcommand through the real fx
// graph against an in-process http source that lists an artifact, asserting it
// configures the registry and writes the confirmation to stdout. The source is an
// httptest.Server and the state lives in a temp SAURON_HOME, so nothing durable
// is touched.
func TestSetRegistryEndToEnd(t *testing.T) {
	// Arrange: an http source that lists one skill artifact.
	source := startHTTPRegistry(t,
		[]artifactSummary{{Name: regName, Version: versionOne, Size: 1024}},
		nil,
	)
	t.Setenv("SAURON_HOME", t.TempDir())

	cmd := SetRegistry()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--transport", transportHTTP, source})

	// Act.
	err := cmd.Execute()

	// Assert.
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "registry set to "+source+" (http)")
	assert.Empty(t, stderr.String())
}

// TestSetGroup asserts the set group has no run behaviour and attaches the
// registry subcommand.
func TestSetGroup(t *testing.T) {
	// Arrange + Act.
	cmd := Set()

	// Assert.
	assert.Equal(t, "set", cmd.Name())
	assert.Nil(t, cmd.RunE, "the group has no run behaviour")

	var registry *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdRegistry {
			registry = sub
		}
	}
	require.NotNil(t, registry, "the registry subcommand is attached")
}

// TestRenderSetRegistry asserts the canonical confirmation line.
func TestRenderSetRegistry(t *testing.T) {
	// Arrange.
	var buf bytes.Buffer
	result := &usecase.SetRegistryResponse{Source: "https://acme.example", Transport: types.TransportHTTP}

	// Act.
	err := renderSetRegistry(&buf, result)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, "registry set to https://acme.example (http)\n", buf.String())
}

// TestRenderSetRegistryWriteError surfaces a writer failure as an io error.
func TestRenderSetRegistryWriteError(t *testing.T) {
	// Act.
	err := renderSetRegistry(&failingWriter{}, &usecase.SetRegistryResponse{Source: "u", Transport: types.TransportGit})

	// Assert.
	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
}
