package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// Shared test-data constants reused across the package's use-case tests (kept
// here to satisfy goconst across files).
const (
	catSortName  = "name"
	catOrderAsc  = "asc"
	catOrderDesc = "desc"
)

// makeSourceFiles returns n fake source.File values for pagination assertions.
// The files share the same name because listAll only inspects len(page).
func makeSourceFiles(n int) []source.File {
	files := make([]source.File, n)
	for i := range files {
		files[i] = fakeFile{name: "item.yaml"}
	}

	return files
}

// TestListAll_Pagination verifies that listAll correctly pages through the source
// until a short page signals exhaustion. Three cases cover the boundaries:
// (a) an empty first page, (b) a short first page, and (c) a full first page
// that triggers a second call whose empty result halts iteration.
func TestListAll_Pagination(t *testing.T) {
	const root = "skills"

	fullPage := makeSourceFiles(int(listPageSize))
	shortPage := makeSourceFiles(10)

	tests := []struct {
		name      string
		setup     func(*source.MockBasedFileSystem)
		wantLen   int
		wantCalls int
	}{
		{
			name: "empty listing terminates after one call",
			setup: func(fs *source.MockBasedFileSystem) {
				fs.On("List", mock.Anything, root, mock.Anything).
					Return(nil, nil).Once()
			},
			wantLen:   0,
			wantCalls: 1,
		},
		{
			name: "short page terminates after one call",
			setup: func(fs *source.MockBasedFileSystem) {
				fs.On("List", mock.Anything, root, mock.Anything).
					Return(shortPage, nil).Once()
			},
			wantLen:   10,
			wantCalls: 1,
		},
		{
			name: "full page triggers second call; empty page terminates",
			setup: func(fs *source.MockBasedFileSystem) {
				fs.On("List", mock.Anything, root, mock.Anything).
					Return(fullPage, nil).Once()
				fs.On("List", mock.Anything, root, mock.Anything).
					Return(nil, nil).Once()
			},
			wantLen:   int(listPageSize),
			wantCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &source.MockBasedFileSystem{}
			tt.setup(fs)

			got, err := listAll(context.Background(), fs, root)

			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			fs.AssertNumberOfCalls(t, "List", tt.wantCalls)
		})
	}
}

func TestInstallPath(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		artifact string
		want     string
	}{
		{
			name:     "skill kind returns skills prefix",
			kind:     types.KindSkill,
			artifact: "go",
			want:     "skills/sauron-go",
		},
		{
			name:     "agent kind returns agents prefix",
			kind:     types.KindAgent,
			artifact: "bar",
			want:     "agents/sauron-bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := installPath(tt.kind, tt.artifact)
			assert.Equal(t, tt.want, got)
		})
	}
}
