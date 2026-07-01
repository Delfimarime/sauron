package api

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

const (
	fileAYAML      = "a.yaml"
	fileBYAML      = "b.yaml"
	fileCYAML      = "c.yaml"
	fileCodeReview = "code-review.yaml"
	fileGoStyle    = "go-style.yaml"
	fileSQLReview  = "sql-review.yaml"
)

// stubFile is a minimal source.File used to exercise the paging helper.
type stubFile struct {
	name string
}

func (f stubFile) Name() string                                { return f.name }
func (f stubFile) IsDirectory() bool                           { return false }
func (f stubFile) Size() int64                                 { return 0 }
func (f stubFile) Version() string                             { return "" }
func (f stubFile) Read(context.Context) (io.ReadCloser, error) { return nil, nil }

// filesOf builds a listing from the given names.
func filesOf(names ...string) []source.File {
	out := make([]source.File, 0, len(names))
	for _, name := range names {
		out = append(out, stubFile{name: name})
	}
	return out
}

// namesOf projects the file names from a listing.
func namesOf(files []source.File) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.Name())
	}
	return out
}

func TestPage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		entries   []source.File
		opts      []source.Option
		wantNames []string
	}{
		{
			name:      "sorts entries ascending by default",
			entries:   filesOf(fileBYAML, fileAYAML),
			wantNames: []string{fileAYAML, fileBYAML},
		},
		{
			name:      "ascending order lists entries by name",
			entries:   filesOf(fileBYAML, fileAYAML),
			opts:      []source.Option{source.WithOrder("asc")},
			wantNames: []string{fileAYAML, fileBYAML},
		},
		{
			name:      "descending order reverses entries",
			entries:   filesOf(fileBYAML, fileAYAML),
			opts:      []source.Option{source.WithOrder("desc")},
			wantNames: []string{fileBYAML, fileAYAML},
		},
		{
			name:      "descending order applies before limit",
			entries:   filesOf(fileAYAML, fileBYAML, fileCYAML),
			opts:      []source.Option{source.WithOrder("desc"), source.WithLimit(1)},
			wantNames: []string{fileCYAML},
		},
		{
			name:      "search filters by case-insensitive substring",
			entries:   filesOf(fileCodeReview, fileGoStyle, fileSQLReview),
			opts:      []source.Option{source.WithSearch("REV")},
			wantNames: []string{fileCodeReview, fileSQLReview},
		},
		{
			name:      "search composes with order, offset and limit",
			entries:   filesOf(fileCodeReview, fileGoStyle, fileSQLReview),
			opts:      []source.Option{source.WithSearch("rev"), source.WithOrder("desc"), source.WithOffset(1), source.WithLimit(1)},
			wantNames: []string{fileCodeReview},
		},
		{
			name:      "empty search matches everything",
			entries:   filesOf(fileAYAML, fileBYAML),
			opts:      []source.Option{source.WithSearch("")},
			wantNames: []string{fileAYAML, fileBYAML},
		},
		{
			name:      "search matching nothing yields empty",
			entries:   filesOf(fileAYAML),
			opts:      []source.Option{source.WithSearch("zzz")},
			wantNames: []string{},
		},
		{
			name:      "limit caps results",
			entries:   filesOf(fileAYAML, fileBYAML, fileCYAML),
			opts:      []source.Option{source.WithLimit(1)},
			wantNames: []string{fileAYAML},
		},
		{
			name:      "offset skips leading entries",
			entries:   filesOf(fileAYAML, fileBYAML),
			opts:      []source.Option{source.WithOffset(1)},
			wantNames: []string{fileBYAML},
		},
		{
			name:      "offset beyond length yields nothing",
			entries:   filesOf(fileAYAML),
			opts:      []source.Option{source.WithOffset(5)},
			wantNames: []string{},
		},
		{
			name:      "empty collection",
			entries:   filesOf(),
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			options := source.Options{}
			for _, opt := range tt.opts {
				opt(&options)
			}

			// Act.
			got := Page(tt.entries, options)

			// Assert.
			assert.Equal(t, tt.wantNames, namesOf(got))
		})
	}
}
