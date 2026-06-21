package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alitto/pond/v2"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
)

// baseYAML and extraYAML are fixture file names used across the git factory tests.
const (
	baseYAML  = "base.yaml"
	extraYAML = "extra.yaml"
)

func TestGitFactory_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []extension.Option
		wantErr error
	}{
		{
			// Ref, SSH key, basic auth, skip-verify and a CA cert are all honored.
			name: "accepts the supported options",
			opts: []extension.Option{
				extension.WithURI("https://example.com/repo.git"),
				extension.WithRef("release"),
				extension.WithSSHKey("/id_ed25519"),
				extension.WithBasicAuth("u", "p"),
				extension.WithSkipTLSVerify(true),
				extension.WithCACert("/ca.pem"),
			},
		},
		{
			// A client certificate cannot be applied to git: usage error.
			name:    "rejects a client certificate",
			opts:    []extension.Option{extension.WithClientCert("/client.crt", "/client.key")},
			wantErr: api.ErrUsage,
		},
		{
			// A lone client key is equally unsupported.
			name:    "rejects a client key alone",
			opts:    []extension.Option{extension.WithClientCert("", "/client.key")},
			wantErr: api.ErrUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			err := newGitFactory(pond.NewPool(1)).Validate(tt.opts...)

			// Assert.
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGitFactory_Open_ChecksOutCommit(t *testing.T) {
	t.Parallel()

	// Arrange: pinning to the default branch's commit hash yields just the base
	// skill; pinning to the release commit yields both — proving the commit is
	// resolved via a full clone after branch and tag lookups miss.
	repo := seedFixtureRepo(t)
	base, release := commitHashes(t, repo)

	tests := []struct {
		name      string
		ref       string
		wantNames []string
	}{
		{name: "base commit", ref: base, wantNames: []string{baseYAML}},
		{name: "release commit", ref: release, wantNames: []string{baseYAML, extraYAML}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			fs, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
				extension.WithURI(repo), extension.WithRef(tt.ref))
			require.NoError(t, err)

			files, listErr := fs.List(context.Background(), ".skills")

			// Assert.
			require.NoError(t, listErr)
			assert.Equal(t, tt.wantNames, names(files))
		})
	}
}

func TestGitFactory_Open_WithCACert(t *testing.T) {
	t.Parallel()

	// Arrange: a readable PEM is loaded into the clone's CA bundle (a local clone
	// ignores it, but the read path is exercised); a missing file is a usage error.
	repo := seedFixtureRepo(t)
	caPath, _ := writeClientCert(t)

	t.Run("readable CA cert is loaded", func(t *testing.T) {
		t.Parallel()

		// Act.
		fs, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
			extension.WithURI(repo), extension.WithCACert(caPath))
		require.NoError(t, err)

		files, listErr := fs.List(context.Background(), ".skills")

		// Assert.
		require.NoError(t, listErr)
		assert.Equal(t, []string{baseYAML}, names(files))
	})

	t.Run("missing CA cert is a usage error", func(t *testing.T) {
		t.Parallel()

		// Act.
		_, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
			extension.WithURI(repo), extension.WithCACert(filepath.Join(t.TempDir(), "absent.pem")))

		// Assert.
		assert.ErrorIs(t, err, api.ErrUsage)
	})
}

func TestGitFactory_Open_TimeoutBoundsClone(t *testing.T) {
	t.Parallel()

	// Arrange.
	repo := seedFixtureRepo(t)

	t.Run("generous timeout still clones", func(t *testing.T) {
		t.Parallel()

		// Act: a positive timeout exercises the bounded-clone path without tripping.
		fs, err := newGitFactory(pond.NewPool(1)).Open(context.Background(),
			extension.WithURI(repo), extension.WithTimeout(30*time.Second))

		// Assert.
		require.NoError(t, err)
		files, listErr := fs.List(context.Background(), ".skills")
		require.NoError(t, listErr)
		assert.Equal(t, []string{baseYAML}, names(files))
	})

	t.Run("a cancelled context fails the clone", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act.
		_, err := newGitFactory(pond.NewPool(1)).Open(ctx, extension.WithURI(repo))

		// Assert.
		assert.ErrorIs(t, err, api.ErrRuntime)
	})
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
			wantNames: []string{baseYAML, extraYAML},
		},
		{
			name:      "tag",
			ref:       "v1",
			wantNames: []string{baseYAML, extraYAML},
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

// commitHashes returns the default-branch and release-branch commit hashes of a
// seeded fixture repository.
func commitHashes(t *testing.T, dir string) (base, release string) {
	t.Helper()

	repo, err := gogit.PlainOpen(dir)
	require.NoError(t, err)

	head, err := repo.Head()
	require.NoError(t, err)

	rel, err := repo.Reference(plumbing.NewBranchReferenceName("release"), true)
	require.NoError(t, err)

	return head.Hash().String(), rel.Hash().String()
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
