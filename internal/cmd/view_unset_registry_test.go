package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
)

// TestRenderUnsetRegistry asserts each outcome renders its canonical report line.
func TestRenderUnsetRegistry(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// outcome is the removal outcome to render.
		outcome usecase.UnsetOutcome
		// want is the exact expected output.
		want string
	}{
		{
			name:    "removed",
			outcome: usecase.UnsetRemoved,
			want:    "registry unset; installed artifacts preserved\n",
		},
		{
			name:    "nothing",
			outcome: usecase.UnsetNothing,
			want:    "no registry configured; nothing was unset\n",
		},
		{
			name:    "preview",
			outcome: usecase.UnsetPreview,
			want:    "registry would be unset; installed artifacts preserved\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := renderUnsetRegistry(&buf, &usecase.UnsetRegistryResponse{Outcome: tt.outcome})

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestRenderUnsetRegistryWriteError surfaces a writer failure as an io error.
func TestRenderUnsetRegistryWriteError(t *testing.T) {
	// Act.
	err := renderUnsetRegistry(&failingWriter{}, &usecase.UnsetRegistryResponse{Outcome: usecase.UnsetRemoved})

	// Assert.
	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
}
