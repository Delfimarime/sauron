package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repeated view-test literals, named to satisfy goconst.
const (
	tblColName = "name"
	tblRowAcme = "acme"
)

// TestPagingLine exercises the applied-paging report: the zero-results line,
// the inclusive from-to range, and the offset math on a single-row window.
func TestPagingLine(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// page, limit, offset, count are pagingLine's inputs.
		page, limit, offset int64
		count               int
		// want is the exact expected line.
		want string
	}{
		{
			name:  "zero count reports zero results",
			page:  9,
			limit: 20,
			count: 0,
			want:  "showing 0 results (page 9, limit 20)",
		},
		{
			name:  "populated window reports the inclusive from-to range",
			page:  1,
			limit: 20,
			count: 2,
			want:  "showing 1–2 (page 1, limit 20)",
		},
		{
			name:   "single-row window reports the inclusive window",
			page:   2,
			limit:  1,
			offset: 1,
			count:  1,
			want:   "showing 2–2 (page 2, limit 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, pagingLine(tt.page, tt.limit, tt.offset, tt.count))
		})
	}
}

// TestTableRender exercises the rendering rules: verbatim headers, aligned
// columns, the empty-cell placeholder, and the no-output-for-zero-rows rule.
func TestTableRender(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// view is the value under test.
		view table
		// want is the exact expected output.
		want string
	}{
		{
			name: "zero rows produce no output",
			view: table{Headers: []string{tblColName, "uri"}},
			want: "",
		},
		{
			name: "headers verbatim and columns aligned",
			view: table{
				Headers: []string{tblColName, "transport", "uri"},
				Rows: [][]string{
					{tblRowAcme, "git", vGitURI},
					{"internal", "http", "https://reg.example.com/"},
				},
			},
			want: "name      transport  uri\n" +
				"acme      git        git@github.com:acme/artifacts.git\n" +
				"internal  http       https://reg.example.com/\n",
		},
		{
			name: "camelCase header renders verbatim, never uppercased",
			view: table{
				Headers: []string{tblColName, "lastUpdatedAt"},
				Rows:    [][]string{{tblRowAcme, "2026-06-21T07:30:00Z"}},
			},
			want: "name  lastUpdatedAt\n" +
				"acme  2026-06-21T07:30:00Z\n",
		},
		{
			name: "empty cell renders the placeholder",
			view: table{
				Headers: []string{tblColName, "ref"},
				Rows:    [][]string{{tblRowAcme, ""}},
			},
			want: "name  ref\n" +
				"acme  —\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := tt.view.render(&buf)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// failingWriter fails after allowing writeAfter successful writes, so the error
// paths of render are reachable.
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
			view := table{
				Headers: []string{tblColName},
				Rows:    [][]string{{tblRowAcme}, {"beta"}},
			}

			// Act.
			err := view.render(&failingWriter{writeAfter: tt.writeAfter})

			// Assert.
			require.Error(t, err)
		})
	}
}

// TestBuildTable exercises the shared slice-to-table composition: one row per
// item via row, under the given headers.
func TestBuildTable(t *testing.T) {
	// Act.
	got := buildTable([]string{tblColName, "kind"}, []string{"a", "b"}, func(s string) []string {
		return []string{s, "skill"}
	})

	// Assert.
	assert.Equal(t, table{
		Headers: []string{tblColName, "kind"},
		Rows:    [][]string{{"a", "skill"}, {"b", "skill"}},
	}, got)
}

// TestBuildTableEmpty confirms an empty slice yields a zero-row table.
func TestBuildTableEmpty(t *testing.T) {
	// Act.
	got := buildTable([]string{tblColName}, []string{}, func(s string) []string { return []string{s} })

	// Assert.
	assert.Equal(t, table{Headers: []string{tblColName}, Rows: [][]string{}}, got)
}

// descriptor-view-test literals, named to satisfy goconst across the package.
const (
	dLabelTransport = "transport"
	dLabelURI       = "uri"
	dLabelAuth      = "auth"
	dLabelUsername  = "username"
	dValGit         = "git"
	dValUserRef     = "${env:ACME_USER}"
)

// TestDescriptorRender exercises the rendering rules: aligned leaf values, the
// nested section block, and the no-output-for-zero-fields rule.
func TestDescriptorRender(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// view is the value under test.
		view descriptor
		// want is the exact expected output.
		want string
	}{
		{
			name: "no fields produce no output",
			view: descriptor{},
			want: "",
		},
		{
			name: "leaf values align to the widest label",
			view: descriptor{Fields: []descriptorField{
				{Label: tblColName, Value: tblRowAcme},
				{Label: dLabelTransport, Value: dValGit},
				{Label: dLabelURI, Value: vGitURI},
			}},
			want: "name:       acme\n" +
				"transport:  git\n" +
				"uri:        git@github.com:acme/artifacts.git\n",
		},
		{
			name: "a section renders its children indented and aligned",
			view: descriptor{Fields: []descriptorField{
				{Label: tblColName, Value: tblRowAcme},
				{Label: dLabelTransport, Value: dValGit},
				{Label: dLabelAuth, Children: []descriptorField{
					{Label: dLabelUsername, Value: dValUserRef},
					{Label: "password", Value: "${env:ACME_TOKEN}"},
				}},
				{Label: describeFieldTimeout, Value: "30s"},
			}},
			want: "name:       acme\n" +
				"transport:  git\n" +
				"auth:\n" +
				"  username: ${env:ACME_USER}\n" +
				"  password: ${env:ACME_TOKEN}\n" +
				"timeout:    30s\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := tt.view.render(&buf)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestDescriptorRenderWriteError surfaces a writer failure rather than swallowing
// it, on both a leaf line and a section header line.
func TestDescriptorRenderWriteError(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// view is the value under test.
		view descriptor
		// writeAfter is the number of successful writes before the failure.
		writeAfter int
	}{
		{
			name:       "leaf line write fails",
			view:       descriptor{Fields: []descriptorField{{Label: "name", Value: "acme"}}},
			writeAfter: 0,
		},
		{
			name: "section header write fails",
			view: descriptor{Fields: []descriptorField{
				{Label: "auth", Children: []descriptorField{{Label: dLabelUsername, Value: "u"}}},
			}},
			writeAfter: 0,
		},
		{
			name: "section child write fails",
			view: descriptor{Fields: []descriptorField{
				{Label: "auth", Children: []descriptorField{{Label: dLabelUsername, Value: "u"}}},
			}},
			writeAfter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := tt.view.render(&failingWriter{writeAfter: tt.writeAfter})

			// Assert.
			require.Error(t, err)
		})
	}
}

// TestSelectFields exercises the shared field-selector directly: an empty
// request yields the full order, an unknown field is a usage error, the first
// field is forced present and first, and duplicates are deduped.
func TestSelectFields(t *testing.T) {
	order := []string{"name", "version", "updatedAt"}

	t.Run("empty request yields the full order", func(t *testing.T) {
		got, err := selectFields(nil, order, "name")

		require.NoError(t, err)
		assert.Equal(t, order, got)
	})

	t.Run("unknown field is a usage error", func(t *testing.T) {
		_, err := selectFields([]string{"bogus"}, order, "name")

		require.Error(t, err)
		assert.ErrorIs(t, err, errInvalidFlag)
	})

	t.Run("first is forced present and first even when omitted", func(t *testing.T) {
		got, err := selectFields([]string{"version"}, order, "name")

		require.NoError(t, err)
		assert.Equal(t, []string{"name", "version"}, got)
	})

	t.Run("duplicates are deduped, order of first occurrence kept", func(t *testing.T) {
		got, err := selectFields([]string{"version", "version", "name"}, order, "name")

		require.NoError(t, err)
		assert.Equal(t, []string{"name", "version"}, got)
	})
}
