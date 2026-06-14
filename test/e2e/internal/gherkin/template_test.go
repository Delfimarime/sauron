//go:build unit

package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func worldForTemplate() *World {
	return &World{
		Environment: map[string]any{
			"App":     map[string]any{"Version": "0.0.0", "Tag": "SNAPSHOT"},
			"GitHash": "abc1234",
		},
		Variables: map[string]any{"HomeDirectory": "/tmp/h"},
	}
}

func TestRender(t *testing.T) {
	out, err := render("v{{.Environment.App.Version}}-{{.Environment.App.Tag}}", worldForTemplate())
	require.NoError(t, err)
	assert.Equal(t, "v0.0.0-SNAPSHOT", out)

	out, err = render("Home: {{.Variables.HomeDirectory}}", worldForTemplate())
	require.NoError(t, err)
	assert.Equal(t, "Home: /tmp/h", out)
}

func TestRenderMissingKeyErrors(t *testing.T) {
	_, err := render("{{.Environment.App.Nope}}", worldForTemplate())
	assert.Error(t, err)
}
