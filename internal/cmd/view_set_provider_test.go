package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/usecase"
)

// TestRenderSetProvider asserts the rendered output for each outcome shape.
func TestRenderSetProvider(t *testing.T) {
	tests := []struct {
		name   string
		result *usecase.SetProviderResult
		want   string
	}{
		{
			name: "change with both groups",
			result: &usecase.SetProviderResult{
				Provider: "zencoder",
				Migrated: 2,
				Skills:   []string{"sauron-acme-go-style"},
				Agents:   []string{"sauron-acme-code-reviewer"},
			},
			want: "skills:\n  ~ sauron-acme-go-style\nagents:\n  ~ sauron-acme-code-reviewer\nprovider set to \"zencoder\"; 2 artifacts migrated\n",
		},
		{
			name: "change with only skills",
			result: &usecase.SetProviderResult{
				Provider: nameClaude,
				Migrated: 1,
				Skills:   []string{"sauron-acme-go-style"},
			},
			want: "skills:\n  ~ sauron-acme-go-style\nprovider set to \"claude\"; 1 artifacts migrated\n",
		},
		{
			name:   "first set with nothing installed",
			result: &usecase.SetProviderResult{Provider: nameClaude},
			want:   "provider set to \"claude\"\n",
		},
		{
			name:   "already active reports no change",
			result: &usecase.SetProviderResult{Provider: nameClaude, Unchanged: true},
			want:   "provider already set to \"claude\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := renderSetProvider(&buf, tt.result)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestRenderSetProviderWriteError surfaces a writer failure as an io error on
// each reachable write.
func TestRenderSetProviderWriteError(t *testing.T) {
	result := &usecase.SetProviderResult{
		Provider: "zencoder",
		Migrated: 2,
		Skills:   []string{"s"},
		Agents:   []string{"a"},
	}

	for _, after := range []int{0, 1, 2, 3, 4} {
		err := renderSetProvider(&failingWriter{writeAfter: after}, result)
		var ucErr *usecase.Error
		require.ErrorAs(t, err, &ucErr)
		assert.Equal(t, usecase.TypeIO, ucErr.Type)
	}
}

// TestRenderSetProviderUnchangedWriteError surfaces the writer failure on the
// no-change path.
func TestRenderSetProviderUnchangedWriteError(t *testing.T) {
	err := renderSetProvider(&failingWriter{}, &usecase.SetProviderResult{Provider: nameClaude, Unchanged: true})
	var ucErr *usecase.Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, usecase.TypeIO, ucErr.Type)
}
