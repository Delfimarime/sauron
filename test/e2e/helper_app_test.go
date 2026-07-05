package e2e

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveFeaturePaths(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"list_skills.feature", "list_agents.feature"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("Feature: x"), 0o600))
	}

	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{
			name: "empty raw returns the whole directory",
			raw:  "",
			want: []string{dir},
		},
		{
			name: "single valid name resolves to a joined path",
			raw:  "list_skills.feature",
			want: []string{filepath.Join(dir, "list_skills.feature")},
		},
		{
			name: "multiple valid names with surrounding whitespace resolve in order",
			raw:  " list_skills.feature , list_agents.feature ",
			want: []string{
				filepath.Join(dir, "list_skills.feature"),
				filepath.Join(dir, "list_agents.feature"),
			},
		},
		{
			name:    "unknown name is a hard error",
			raw:     "does_not_exist.feature",
			wantErr: true,
		},
		{
			name:    "a raw string resolving to nothing is an error",
			raw:     ",, ,",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveFeaturePaths(dir, tt.raw)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveConcurrency(t *testing.T) {
	t.Run("explicit positive value is returned as-is, even above the cap", func(t *testing.T) {
		assert.Equal(t, 1, resolveConcurrency(1))
		assert.Equal(t, 42, resolveConcurrency(42))
	})

	t.Run("explicit zero falls back to the CPU-aware default", func(t *testing.T) {
		got := resolveConcurrency(0)
		assert.GreaterOrEqual(t, got, 1)
		assert.LessOrEqual(t, got, 5)
	})

	t.Run("explicit negative falls back to the CPU-aware default", func(t *testing.T) {
		got := resolveConcurrency(-1)
		assert.GreaterOrEqual(t, got, 1)
		assert.LessOrEqual(t, got, 5)
	})

	t.Run("default matches the documented formula", func(t *testing.T) {
		want := runtime.NumCPU() / 2
		if want < 1 {
			want = 1
		}
		if want > 5 {
			want = 5
		}
		assert.Equal(t, want, resolveConcurrency(0))
	})
}
