package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alitto/pond/v2"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
)

// baseYAML is a fixture file name used across the git factory tests.
const baseYAML = "base.yaml"

func TestGitFactory_Validate_AcceptsEverything(t *testing.T) {
	t.Parallel()

	// Act.
	err := newGitFactory(pond.NewPool(1)).Validate(
		extension.WithURI("https://example.com/repo.git"),
		extension.WithRef("release"),
		extension.WithSSHKey("/id_ed25519"),
		extension.WithBasicAuth("u", "p"),
		extension.WithSkipTLSVerify(true),
	)

	// Assert.
	assert.NoError(t, err)
}

func TestGitFactory_Open_ChecksOutRef(t *testing.T) {
	t.Parallel()

	// Arrange: a fixture repo whose default branch holds a single skill and a
	// "release" branch that adds a second one.
	repo := seedFixtureRepo(t)

	tests := []struct {
		name      string
		ref       string
		wantNames []string
	}{
		{
			name:      "default branch",
			ref:       "",
			wantNames: []string{baseYAML},
		},
		{
			name:      "named branch",
			ref:       "release",
			wantNames: []string{baseYAML, "extra.yaml"},
		},
		{
			name:      "tag",
			ref:       "v1",
			wantNames: []string{baseYAML, "extra.yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			opts := []extension.Option{extension.WithURI(repo)}
			if tt.ref != "" {
				opts = append(opts, extension.WithRef(tt.ref))
			}
			fs, err := newGitFactory(pond.NewPool(1)).Open(context.Background(), opts...)
			require.NoError(t, err)

			files, listErr := fs.List(context.Background(), ".skills")

			// Assert.
			require.NoError(t, listErr)
			assert.Equal(t, tt.wantNames, names(files))
		})
	}
}

func TestGitFactory_Open_WithBasicAuth(t *testing.T) {
	t.Parallel()

	// Arrange: basic auth is resolved into go-git credentials; a local clone
	// ignores them but the resolution path is still exercised.
	repo := seedFixtureRepo(t)

	// Act.
	fs, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
		extension.WithURI(repo), extension.WithBasicAuth("u", "p"))
	require.NoError(t, err)

	files, listErr := fs.List(context.Background(), ".skills")

	// Assert.
	require.NoError(t, listErr)
	assert.Equal(t, []string{baseYAML}, names(files))
}

func TestGitFactory_Open_UnknownRef(t *testing.T) {
	t.Parallel()

	// Arrange.
	repo := seedFixtureRepo(t)

	// Act.
	_, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
		extension.WithURI(repo), extension.WithRef("does-not-exist"))

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrRuntime)
}

func TestGitFactory_Open_UnreachableRepo(t *testing.T) {
	t.Parallel()

	// Act.
	_, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
		extension.WithURI(filepath.Join(t.TempDir(), "absent.git")))

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrRuntime)
}

func TestGitFactory_Open_BadSSHKey(t *testing.T) {
	t.Parallel()

	// Arrange.
	key := filepath.Join(t.TempDir(), "id")
	require.NoError(t, os.WriteFile(key, []byte("not a key"), 0o600))

	// Act.
	_, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
		extension.WithURI("ssh://git@example.com/repo.git"), extension.WithSSHKey(key))

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrUsage)
}

// seedFixtureRepo builds an on-disk repository with a default branch, a
// "release" branch, and a "v1" tag, returning its path for cloning.
func seedFixtureRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)

	tree, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".skills"), 0o755))
	writeAndStage(t, tree, dir, ".skills/base.yaml", "base")
	base := commit(t, tree)

	head, err := repo.Head()
	require.NoError(t, err)

	release := plumbing.NewBranchReferenceName("release")
	require.NoError(t, repo.Storer.SetReference(plumbing.NewHashReference(release, base)))
	require.NoError(t, tree.Checkout(&gogit.CheckoutOptions{Branch: release}))

	writeAndStage(t, tree, dir, ".skills/extra.yaml", "extra")
	releaseCommit := commit(t, tree)

	_, err = repo.CreateTag("v1", releaseCommit, nil)
	require.NoError(t, err)

	require.NoError(t, tree.Checkout(&gogit.CheckoutOptions{Branch: head.Name()}))

	return dir
}

// writeAndStage writes a file under the worktree and stages it.
func writeAndStage(t *testing.T, tree *gogit.Worktree, dir, rel, body string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, rel)), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644))
	_, err := tree.Add(rel)
	require.NoError(t, err)
}

// commit records a commit and returns its hash.
func commit(t *testing.T, tree *gogit.Worktree) plumbing.Hash {
	t.Helper()

	hash, err := tree.Commit("seed", &gogit.CommitOptions{
		Author: &object.Signature{Name: "fixture", Email: "fixture@example.com"},
	})
	require.NoError(t, err)

	return hash
}
