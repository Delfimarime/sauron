package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetProviderHandlerRejectsUnknownName asserts an unsupported provider name
// surfaces as a usage error mapped to exit code 2.
func TestSetProviderHandlerRejectsUnknownName(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", t.TempDir())

	// Act.
	err := setProvider(context.Background(), []string{nameBogus}, &bytes.Buffer{})

	// Assert.
	require.Error(t, err)
	assert.Equal(t, exitUsage, exitCode(err))
}

// TestSetProviderCommand exercises the assembled subcommand: ExactArgs(1) and the
// unsupported-name path, all mapped to the usage exit code. A temporary
// SAURON_HOME keeps nothing durable.
func TestSetProviderCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "rejects a missing argument", args: []string{}},
		{name: caseUnexpectedArg, args: []string{nameClaude, "extra"}},
		{name: "rejects an unsupported name", args: []string{nameBogus}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			t.Setenv("SAURON_HOME", t.TempDir())
			cmd := SetProvider()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(tt.args)
			cmd.SetContext(context.Background())

			// Act.
			err := cmd.Execute()

			// Assert.
			require.Error(t, err)
			assert.Equal(t, exitUsage, exitCode(err))
		})
	}
}

// TestSetProviderEndToEnd drives the assembled subcommand through the real fx
// graph: the first invocation records the provider, a second invocation of the
// same provider reports no change. State lives in a temp SAURON_HOME.
func TestSetProviderEndToEnd(t *testing.T) {
	// Arrange.
	t.Setenv("SAURON_HOME", t.TempDir())

	// Act: first set records the provider.
	var first bytes.Buffer
	set := SetProvider()
	set.SetOut(&first)
	set.SetErr(&bytes.Buffer{})
	set.SetContext(context.Background())
	set.SetArgs([]string{nameClaude})
	require.NoError(t, set.Execute())

	// Assert: with nothing installed the summary reads cleanly.
	assert.Equal(t, "provider set to \"claude\"\n", first.String())

	// Act: re-setting the active provider changes nothing.
	var second bytes.Buffer
	reset := SetProvider()
	reset.SetOut(&second)
	reset.SetErr(&bytes.Buffer{})
	reset.SetContext(context.Background())
	reset.SetArgs([]string{nameClaude})
	require.NoError(t, reset.Execute())

	// Assert.
	assert.Equal(t, "provider already set to \"claude\"\n", second.String())
}

// TestSetGroupAttachesProvider asserts the provider subcommand is attached to the
// set group.
func TestSetGroupAttachesProvider(t *testing.T) {
	// Arrange + Act.
	cmd := Set()

	// Assert.
	var provider bool
	for _, sub := range cmd.Commands() {
		if sub.Name() == subcmdProvider {
			provider = true
		}
	}
	assert.True(t, provider, "the provider subcommand is attached")
}
