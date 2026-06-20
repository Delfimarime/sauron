package api

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

const (
	fileAYAML = "a.yaml"
	fileBYAML = "b.yaml"
	dirSkills = ".skills"
)

// seed returns a memory filesystem holding a .skills directory with the given
// entries.
func seed(t *testing.T, entries map[string][]byte) afero.Fs {
	t.Helper()

	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(dirSkills, 0o755))
	for name, body := range entries {
		require.NoError(t, afero.WriteFile(fs, dirSkills+"/"+name, body, 0o644))
	}

	return fs
}

func TestDirectory_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		entries   map[string][]byte
		uri       string
		opts      []source.Option
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "lists sorted entries",
			entries:   map[string][]byte{fileBYAML: []byte("b"), fileAYAML: []byte("aa")},
			uri:       dirSkills,
			wantNames: []string{fileAYAML, fileBYAML},
		},
		{
			name:      "limit caps results",
			entries:   map[string][]byte{fileAYAML: []byte("a"), fileBYAML: []byte("b"), "c.yaml": []byte("c")},
			uri:       dirSkills,
			opts:      []source.Option{source.WithLimit(1)},
			wantNames: []string{fileAYAML},
		},
		{
			name:      "offset skips leading entries",
			entries:   map[string][]byte{fileAYAML: []byte("a"), fileBYAML: []byte("b")},
			uri:       dirSkills,
			opts:      []source.Option{source.WithOffset(1)},
			wantNames: []string{fileBYAML},
		},
		{
			name:      "offset beyond length yields nothing",
			entries:   map[string][]byte{fileAYAML: []byte("a")},
			uri:       dirSkills,
			opts:      []source.Option{source.WithOffset(5)},
			wantNames: []string{},
		},
		{
			name:      "empty collection",
			entries:   map[string][]byte{},
			uri:       dirSkills,
			wantNames: []string{},
		},
		{
			name:    "missing directory is a runtime error",
			entries: map[string][]byte{},
			uri:     ".agents",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			dir := NewDirectory(seed(t, tt.entries))

			// Act.
			files, err := dir.List(context.Background(), tt.uri, tt.opts...)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrRuntime)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantNames, names(files))
		})
	}
}

func TestDirectory_List_StatMetadata(t *testing.T) {
	t.Parallel()

	// Arrange.
	dir := NewDirectory(seed(t, map[string][]byte{fileAYAML: []byte("hello")}))

	// Act.
	files, err := dir.List(context.Background(), dirSkills)

	// Assert.
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, fileAYAML, files[0].Name())
	assert.False(t, files[0].IsDirectory())
	assert.Equal(t, int64(5), files[0].Size())
	assert.Equal(t, "", files[0].Version())
}

func TestDirectory_DescribeAndGetAreNotImplemented(t *testing.T) {
	t.Parallel()

	// Arrange.
	dir := NewDirectory(afero.NewMemMapFs())

	// Act.
	_, describeErr := dir.Describe(context.Background(), dirSkills)
	_, getErr := dir.Get(context.Background(), dirSkills+"/"+fileAYAML)

	// Assert.
	assert.ErrorIs(t, describeErr, source.ErrNotImplemented)
	assert.ErrorIs(t, getErr, source.ErrNotImplemented)
}

func TestDirectoryEntry_ReadIsNotImplemented(t *testing.T) {
	t.Parallel()

	// Arrange.
	dir := NewDirectory(seed(t, map[string][]byte{fileAYAML: []byte("x")}))
	files, err := dir.List(context.Background(), dirSkills)
	require.NoError(t, err)
	require.Len(t, files, 1)

	// Act.
	_, readErr := files[0].Read(context.Background())

	// Assert.
	assert.ErrorIs(t, readErr, source.ErrNotImplemented)
}

// names projects the file names from a listing.
func names(files []source.File) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.Name())
	}
	return out
}
