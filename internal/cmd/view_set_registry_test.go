package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

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
