package presentation

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repeated test literals, named to satisfy goconst.
const (
	colName = "name"
	rowAcme = "acme"
)

// TestTableRender exercises the rendering rules: uppercase headers, aligned
// columns, the empty-cell placeholder, and the no-output-for-zero-rows rule.
func TestTableRender(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// table is the value under test.
		table Table
		// want is the exact expected output.
		want string
	}{
		{
			name:  "zero rows produce no output",
			table: Table{Headers: []string{colName, "uri"}},
			want:  "",
		},
		{
			name: "headers upper-cased and columns aligned",
			table: Table{
				Headers: []string{colName, "transport", "uri"},
				Rows: [][]string{
					{rowAcme, "git", "git@github.com:acme/artifacts.git"},
					{"internal", "http", "https://reg.example.com/"},
				},
			},
			want: "NAME      TRANSPORT  URI\n" +
				"acme      git        git@github.com:acme/artifacts.git\n" +
				"internal  http       https://reg.example.com/\n",
		},
		{
			name: "empty cell renders the placeholder",
			table: Table{
				Headers: []string{colName, "ref"},
				Rows:    [][]string{{rowAcme, ""}},
			},
			want: "NAME  REF\n" +
				"acme  —\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := tt.table.Render(&buf)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// failingWriter fails after allowing writeAfter successful writes, so the error
// paths of Render are reachable.
type failingWriter struct {
	writeAfter int
	writes     int
}

func (f *failingWriter) Write(p []byte) (int, error) {
	f.writes++
	if f.writes > f.writeAfter {
		return 0, errors.New("write failed")
	}
	return len(p), nil
}

// TestTableRenderWriteError surfaces a writer failure rather than swallowing it.
func TestTableRenderWriteError(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// writeAfter is the number of successful writes before the failure.
		writeAfter int
	}{
		{name: "header write fails", writeAfter: 0},
		{name: "row write fails", writeAfter: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			table := Table{
				Headers: []string{colName},
				Rows:    [][]string{{rowAcme}, {"beta"}},
			}

			// Act.
			err := table.Render(&failingWriter{writeAfter: tt.writeAfter})

			// Assert.
			require.Error(t, err)
		})
	}
}
