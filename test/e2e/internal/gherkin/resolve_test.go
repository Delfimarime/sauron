//go:build unit

package gherkin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueOfPlainValues(t *testing.T) {
	ctx := context.Background()
	rt := &fakeRuntime{}

	s, err := valueOf[string](ctx, rt, "literal")
	require.NoError(t, err)
	assert.Equal(t, "literal", s)

	n, err := valueOf[int](ctx, rt, "42")
	require.NoError(t, err)
	assert.Equal(t, 42, n)

	_, err = valueOf[int](ctx, rt, "not-a-number")
	assert.Error(t, err)
}

func TestValueOfReferences(t *testing.T) {
	ctx := context.Background()
	rt := &fakeRuntime{
		folders:    map[string]*fakeSource{"default": {path: "/tmp/reg"}},
		webservers: map[string]*fakeSource{"acme": {url: "http://registry-http-acme"}},
	}

	tests := map[string]struct {
		expr string
		want string
	}{
		"folder default (implicit alias)": {"#{.folder.path}", "/tmp/reg"},
		"folder default (explicit alias)": {"#{.folder.default.path}", "/tmp/reg"},
		"webserver aliased":               {"#{.webserver.acme.url}", "http://registry-http-acme"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := valueOf[string](ctx, rt, tc.expr)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestValueOfReferenceErrors(t *testing.T) {
	ctx := context.Background()
	rt := &fakeRuntime{}

	tests := map[string]string{
		"unknown capability":    "#{.pod.default.name}",
		"wrong attr for folder": "#{.folder.default.url}",
		"wrong attr for web":    "#{.webserver.default.path}",
		"missing leading dot":   "#{folder.path}",
		"too many segments":     "#{.folder.default.deep.path}",
		"git is deferred":       "#{.git.default.url}",
	}
	for name, expr := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := valueOf[string](ctx, rt, expr)
			assert.Error(t, err)
		})
	}
}

func TestParseReferenceDefaultsAlias(t *testing.T) {
	ref, err := parseReference("#{.webserver.url}")
	require.NoError(t, err)
	assert.Equal(t, "webserver", ref.capability)
	assert.Equal(t, defaultAlias, ref.alias)
	assert.Equal(t, "url", ref.attr)
}
