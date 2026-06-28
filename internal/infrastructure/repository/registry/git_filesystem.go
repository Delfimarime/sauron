package registry

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	httpauth "github.com/go-git/go-git/v5/plumbing/transport/http"
	sshauth "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// gitFactory opens a git-backed source by cloning a repository into memory.
type gitFactory struct{}

// newGitFactory builds a gitFactory.
func newGitFactory() *gitFactory {
	return &gitFactory{}
}

// Validate rejects a client certificate, which the git transport cannot apply,
// and accepts every other option. The configuration is verified against the
// remote only when the source is opened.
func (gitFactory) Validate(opts ...extension.Option) error {
	options := api.Resolve(opts)
	if options.ClientCert != "" || options.ClientKey != "" {
		return fmt.Errorf("%w: a client certificate is not supported", api.ErrUsage)
	}

	return nil
}

// Open clones the repository at the URI into go-git's in-memory object store,
// resolves the requested reference, and returns a read-only view over the
// resulting commit tree. A positive timeout bounds the clone alone.
func (f gitFactory) Open(ctx context.Context, opts ...extension.Option) (source.FileSystem, error) {
	options := api.Resolve(opts)

	if err := f.Validate(opts...); err != nil {
		return nil, err
	}

	cloneCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		cloneCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	repo, root, err := f.clone(cloneCtx, options)
	if err != nil {
		return nil, err
	}

	return newGitTreeSource(repo, root), nil
}

// clone resolves options.Ref against the remote: an empty ref clones the default
// branch; otherwise the ref is tried as a branch, then a tag, then a commit. It
// returns the repository and the resolved commit's tree.
func (f gitFactory) clone(ctx context.Context, options extension.Options) (*gogit.Repository, *object.Tree, error) {
	base, err := f.cloneOptions(options)
	if err != nil {
		return nil, nil, err
	}

	if options.Ref == "" {
		return f.cloneRef(ctx, base, plumbing.ReferenceName(""))
	}

	for _, ref := range []plumbing.ReferenceName{
		plumbing.NewBranchReferenceName(options.Ref),
		plumbing.NewTagReferenceName(options.Ref),
	} {
		if repo, tree, err := f.cloneRef(ctx, base, ref); err == nil {
			return repo, tree, nil
		}
	}

	return f.cloneCommit(ctx, base, options.Ref)
}

// cloneOptions assembles the shared clone configuration: credentials, a shallow
// depth, and the TLS posture the git transport honors.
//
// ponytail: the clone is held in RAM (go-git memory store + memfs). A shallow
// depth-1 clone of a text registry is small; move to a bounded on-disk clone
// only if a registry is ever large enough to threaten the process memory
// ceiling.
func (f gitFactory) cloneOptions(options extension.Options) (gogit.CloneOptions, error) {
	auth, err := f.authFrom(options)
	if err != nil {
		return gogit.CloneOptions{}, err
	}

	opts := gogit.CloneOptions{
		URL:             options.URI,
		Auth:            auth,
		Depth:           1,
		InsecureSkipTLS: options.SkipTLSVerify,
	}

	if options.CACert != "" {
		bundle, err := os.ReadFile(options.CACert) // #nosec G304 -- certificate path is operator-supplied
		if err != nil {
			return gogit.CloneOptions{}, fmt.Errorf("%w: read CA certificate: %w", api.ErrUsage, err)
		}
		opts.CABundle = bundle
	}

	return opts, nil
}

// cloneRef performs a single shallow in-memory clone, pinned to ref when one is
// given, and returns the repository and its HEAD commit tree.
func (f gitFactory) cloneRef(ctx context.Context, base gogit.CloneOptions, ref plumbing.ReferenceName) (*gogit.Repository, *object.Tree, error) {
	opts := base
	if ref != "" {
		opts.ReferenceName = ref
		opts.SingleBranch = true
	}

	repo, err := gogit.CloneContext(ctx, memory.NewStorage(), memfs.New(), &opts)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: clone %q: %w", api.ErrRuntime, opts.URL, err)
	}

	tree, err := headTree(repo)
	if err != nil {
		return nil, nil, err
	}

	return repo, tree, nil
}

// cloneCommit performs a full in-memory clone and resolves an arbitrary commit —
// the only ref kind a shallow single-branch clone cannot reach — returning the
// repository and that commit's tree.
func (f gitFactory) cloneCommit(ctx context.Context, base gogit.CloneOptions, ref string) (*gogit.Repository, *object.Tree, error) {
	opts := base
	opts.Depth = 0
	opts.ReferenceName = ""
	opts.SingleBranch = false

	repo, err := gogit.CloneContext(ctx, memory.NewStorage(), memfs.New(), &opts)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: clone %q: %w", api.ErrRuntime, opts.URL, err)
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: resolve ref %q: %w", api.ErrRuntime, ref, err)
	}

	tree, err := treeOf(repo, *hash)
	if err != nil {
		return nil, nil, err
	}

	return repo, tree, nil
}

// headTree returns the tree of repo's HEAD commit.
func headTree(repo *gogit.Repository) (*object.Tree, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("%w: resolve HEAD: %w", api.ErrRuntime, err)
	}

	return treeOf(repo, head.Hash())
}

// treeOf returns the tree of the commit identified by hash.
func treeOf(repo *gogit.Repository, hash plumbing.Hash) (*object.Tree, error) {
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve commit %q: %w", api.ErrRuntime, hash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("%w: resolve tree %q: %w", api.ErrRuntime, hash, err)
	}

	return tree, nil
}

// authFrom resolves the clone credentials from options, preferring an SSH key.
func (gitFactory) authFrom(options extension.Options) (transport.AuthMethod, error) {
	if options.SSHKey != "" {
		auth, err := sshauth.NewPublicKeysFromFile("git", options.SSHKey, options.Password)
		if err != nil {
			return nil, fmt.Errorf("%w: load ssh key: %w", api.ErrUsage, err)
		}
		return auth, nil
	}

	if options.Username != "" || options.Password != "" {
		return &httpauth.BasicAuth{Username: options.Username, Password: options.Password}, nil
	}

	return nil, nil
}
