package registry

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// baseYAML and extraYAML are fixture file names used across the git factory
// tests; skillGo names an artifact directory the tree-source tests assert over.
// rootSkills and rootAgents are defined in roots.go.
const (
	baseYAML   = "base.yaml"
	extraYAML  = "extra.yaml"
	skillGo    = "sauron-go"
	artifactGo = rootSkills + "/" + skillGo
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
			err := newGitFactory().Validate(tt.opts...)

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
			fs, err := newGitFactory().Open(context.Background(),
				extension.WithURI(repo), extension.WithRef(tt.ref))
			require.NoError(t, err)

			files, listErr := fs.List(context.Background(), rootSkills)

			// Assert.
			require.NoError(t, listErr)
			assert.Equal(t, tt.wantNames, names(files))
		})
	}
}

func TestGitFactory_Open_CommitSHA_VersionIsTreeHash(t *testing.T) {
	t.Parallel()

	// Arrange: opening via a raw commit SHA forces the cloneCommit path, which is
	// distinct from cloneRef. Verify that the artifact entry's Version() equals
	// the independently computed tree-object hash at that commit.
	repo := seedArtifactRepo(t)

	onDisk, err := gogit.PlainOpen(repo)
	require.NoError(t, err)
	head, err := onDisk.Head()
	require.NoError(t, err)
	commitSHA := head.Hash().String()

	want := subtreeHash(t, repo, artifactGo)

	// Act: branch/tag lookups miss a raw SHA, so cloneCommit is reached.
	fs, openErr := newGitFactory().Open(context.Background(),
		extension.WithURI(repo), extension.WithRef(commitSHA))
	require.NoError(t, openErr)

	listing, listErr := fs.List(context.Background(), rootSkills)
	require.NoError(t, listErr)

	// Assert: the artifact directory entry carries the subtree's tree-object hash.
	var entry source.File
	for _, f := range listing {
		if f.Name() == skillGo {
			entry = f
		}
	}
	require.NotNil(t, entry, "expected %q in listing", skillGo)
	assert.Equal(t, want, entry.Version())
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
		fs, err := newGitFactory().Open(context.Background(),
			extension.WithURI(repo), extension.WithCACert(caPath))
		require.NoError(t, err)

		files, listErr := fs.List(context.Background(), rootSkills)

		// Assert.
		require.NoError(t, listErr)
		assert.Equal(t, []string{baseYAML}, names(files))
	})

	t.Run("missing CA cert is a usage error", func(t *testing.T) {
		t.Parallel()

		// Act.
		_, err := newGitFactory().Open(context.Background(),
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
		fs, err := newGitFactory().Open(context.Background(),
			extension.WithURI(repo), extension.WithTimeout(30*time.Second))

		// Assert.
		require.NoError(t, err)
		files, listErr := fs.List(context.Background(), rootSkills)
		require.NoError(t, listErr)
		assert.Equal(t, []string{baseYAML}, names(files))
	})

	t.Run("a cancelled context fails the clone", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act.
		_, err := newGitFactory().Open(ctx, extension.WithURI(repo))

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
			fs, err := newGitFactory().Open(context.Background(), opts...)
			require.NoError(t, err)

			files, listErr := fs.List(context.Background(), rootSkills)

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
	fs, err := newGitFactory().Open(context.Background(),
		extension.WithURI(repo), extension.WithBasicAuth("u", "p"))
	require.NoError(t, err)

	files, listErr := fs.List(context.Background(), rootSkills)

	// Assert.
	require.NoError(t, listErr)
	assert.Equal(t, []string{baseYAML}, names(files))
}

func TestGitFactory_Open_UnknownRef(t *testing.T) {
	t.Parallel()

	// Arrange.
	repo := seedFixtureRepo(t)

	// Act.
	_, err := newGitFactory().Open(context.Background(),
		extension.WithURI(repo), extension.WithRef("does-not-exist"))

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrRuntime)
}

func TestGitFactory_Open_UnreachableRepo(t *testing.T) {
	t.Parallel()

	// Act.
	_, err := newGitFactory().Open(context.Background(),
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
	_, err := newGitFactory().Open(context.Background(),
		extension.WithURI("ssh://git@example.com/repo.git"), extension.WithSSHKey(key))

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrUsage)
}

func TestGitFactory_Open_ListArtifacts(t *testing.T) {
	t.Parallel()

	// Arrange: a registry whose skills root holds three artifact directories.
	repo := seedArtifactRepo(t)

	tests := []struct {
		name      string
		uri       string
		opts      []source.Option
		wantNames []string
	}{
		{
			name:      "lists immediate artifact directories sorted",
			uri:       rootSkills,
			wantNames: []string{skillGo, "sauron-python", "sauron-rust"},
		},
		{
			name:      "search filters artifacts",
			uri:       rootSkills,
			opts:      []source.Option{source.WithSearch("go")},
			wantNames: []string{skillGo},
		},
		{
			name:      "descending order applies before limit",
			uri:       rootSkills,
			opts:      []source.Option{source.WithOrder("desc"), source.WithLimit(1)},
			wantNames: []string{"sauron-rust"},
		},
		{
			name:      "lists the agents root",
			uri:       rootAgents,
			wantNames: []string{"sauron-review"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			fs, err := newGitFactory().Open(context.Background(), extension.WithURI(repo))
			require.NoError(t, err)

			files, listErr := fs.List(context.Background(), tt.uri, tt.opts...)

			// Assert.
			require.NoError(t, listErr)
			assert.Equal(t, tt.wantNames, names(files))
			for _, f := range files {
				assert.True(t, f.IsDirectory())
			}
		})
	}
}

func TestGitFactory_Open_ListMissingDirectory(t *testing.T) {
	t.Parallel()

	// Arrange.
	repo := seedArtifactRepo(t)

	// Act.
	fs, err := newGitFactory().Open(context.Background(), extension.WithURI(repo))
	require.NoError(t, err)

	_, listErr := fs.List(context.Background(), "personas")

	// Assert.
	require.Error(t, listErr)
	assert.ErrorIs(t, listErr, api.ErrRuntime)
}

func TestGitFactory_Open_FetchTree(t *testing.T) {
	t.Parallel()

	// Arrange: the go skill is a directory with a nested file.
	repo := seedArtifactRepo(t)

	// Act.
	fs, err := newGitFactory().Open(context.Background(), extension.WithURI(repo))
	require.NoError(t, err)

	files, fetchErr := fs.Fetch(context.Background(), artifactGo)

	// Assert: every blob is returned with an artifact-relative name.
	require.NoError(t, fetchErr)
	assert.ElementsMatch(t, []string{skillMD, "lib/util.go"}, names(files))

	byName := make(map[string]source.File, len(files))
	for _, f := range files {
		byName[f.Name()] = f
	}

	nested := byName["lib/util.go"]
	assert.False(t, nested.IsDirectory())
	assert.Equal(t, int64(len("package lib")), nested.Size())

	reader, readErr := nested.Read(context.Background())
	require.NoError(t, readErr)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	assert.Equal(t, "package lib", string(body))
}

func TestGitFactory_Open_FetchMissingArtifact(t *testing.T) {
	t.Parallel()

	// Arrange.
	repo := seedArtifactRepo(t)

	// Act.
	fs, err := newGitFactory().Open(context.Background(), extension.WithURI(repo))
	require.NoError(t, err)

	_, fetchErr := fs.Fetch(context.Background(), "skills/absent")

	// Assert.
	require.Error(t, fetchErr)
	assert.ErrorIs(t, fetchErr, api.ErrRuntime)
}

func TestGitFactory_Open_DescribeAndGetNotImplemented(t *testing.T) {
	t.Parallel()

	// Arrange.
	repo := seedArtifactRepo(t)
	fs, err := newGitFactory().Open(context.Background(), extension.WithURI(repo))
	require.NoError(t, err)

	// Act.
	_, describeErr := fs.Describe(context.Background(), rootSkills)
	_, getErr := fs.Get(context.Background(), artifactGo)

	// Assert.
	assert.ErrorIs(t, describeErr, source.ErrNotImplemented)
	assert.ErrorIs(t, getErr, source.ErrNotImplemented)
}

func TestGitFactory_Open_ReadDirectoryEntryErrors(t *testing.T) {
	t.Parallel()

	// Arrange: a listing of the skills root yields artifact directories, which
	// are not themselves readable.
	repo := seedArtifactRepo(t)
	fs, err := newGitFactory().Open(context.Background(), extension.WithURI(repo))
	require.NoError(t, err)

	files, listErr := fs.List(context.Background(), rootSkills)
	require.NoError(t, listErr)
	require.NotEmpty(t, files)
	require.True(t, files[0].IsDirectory())

	// Act.
	_, readErr := files[0].Read(context.Background())

	// Assert.
	require.Error(t, readErr)
	assert.ErrorIs(t, readErr, api.ErrRuntime)
}

func TestGitFactory_Open_VersionIsTreeHash(t *testing.T) {
	t.Parallel()

	// Arrange: the native version of an artifact is its directory's git
	// tree-object hash, shared by the listing entry and every fetched blob.
	repo := seedArtifactRepo(t)
	want := subtreeHash(t, repo, artifactGo)

	fs, err := newGitFactory().Open(context.Background(), extension.WithURI(repo))
	require.NoError(t, err)

	// Act.
	listing, listErr := fs.List(context.Background(), rootSkills)
	require.NoError(t, listErr)
	fetched, fetchErr := fs.Fetch(context.Background(), artifactGo)
	require.NoError(t, fetchErr)

	// Assert.
	var entry source.File
	for _, f := range listing {
		if f.Name() == skillGo {
			entry = f
		}
	}
	require.NotNil(t, entry)
	assert.Equal(t, want, entry.Version())
	for _, f := range fetched {
		assert.Equal(t, want, f.Version())
	}
}

// seedArtifactRepo builds an on-disk repository whose skills and agents roots
// hold artifact directories, returning its path for cloning.
func seedArtifactRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)

	tree, err := repo.Worktree()
	require.NoError(t, err)

	for rel, body := range map[string]string{
		"skills/sauron-go/SKILL.md":     "go skill",
		"skills/sauron-go/lib/util.go":  "package lib",
		"skills/sauron-rust/SKILL.md":   "rust skill",
		"skills/sauron-python/SKILL.md": "python skill",
		"agents/sauron-review/AGENT.md": "review agent",
	} {
		writeAndStage(t, tree, dir, rel, body)
	}
	commit(t, tree)

	return dir
}

// subtreeHash returns the git tree-object hash of the subtree at treePath in a
// seeded fixture repository's HEAD commit.
func subtreeHash(t *testing.T, dir, treePath string) string {
	t.Helper()

	repo, err := gogit.PlainOpen(dir)
	require.NoError(t, err)

	head, err := repo.Head()
	require.NoError(t, err)

	commit, err := repo.CommitObject(head.Hash())
	require.NoError(t, err)

	root, err := commit.Tree()
	require.NoError(t, err)

	sub, err := root.Tree(treePath)
	require.NoError(t, err)

	return sub.Hash.String()
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

	require.NoError(t, os.MkdirAll(filepath.Join(dir, rootSkills), 0o755))
	writeAndStage(t, tree, dir, "skills/base.yaml", "base")
	base := commit(t, tree)

	head, err := repo.Head()
	require.NoError(t, err)

	release := plumbing.NewBranchReferenceName("release")
	require.NoError(t, repo.Storer.SetReference(plumbing.NewHashReference(release, base)))
	require.NoError(t, tree.Checkout(&gogit.CheckoutOptions{Branch: release}))

	writeAndStage(t, tree, dir, "skills/extra.yaml", "extra")
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
