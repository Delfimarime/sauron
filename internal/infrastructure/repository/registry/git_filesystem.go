package registry

import (
	"context"
	"fmt"
	"os"

	"github.com/alitto/pond/v2"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	httpauth "github.com/go-git/go-git/v5/plumbing/transport/http"
	sshauth "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/spf13/afero"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// gitFactory opens a git-backed source by cloning a repository.
type gitFactory struct {
	pool pond.Pool
}

// newGitFactory builds a gitFactory that reclaims clones through the given pool.
func newGitFactory(pool pond.Pool) *gitFactory {
	return &gitFactory{pool: pool}
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

// Open clones the repository at the URI into a temporary directory bound to ctx's
// lifetime, resolves the requested reference, and returns a read-only view rooted
// at the clone. A positive timeout bounds the clone alone.
func (f gitFactory) Open(ctx context.Context, opts ...extension.Option) (source.FileSystem, error) {
	options := api.Resolve(opts)

	if err := f.Validate(opts...); err != nil {
		return nil, err
	}

	dir, err := os.MkdirTemp("", "sauron-registry-git-")
	if err != nil {
		return nil, fmt.Errorf("%w: stage clone: %w", api.ErrRuntime, err)
	}
	f.cleanupWhenDone(ctx, dir)

	cloneCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		cloneCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	if err := f.clone(cloneCtx, dir, options); err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}

	return api.NewDirectory(afero.NewBasePathFs(afero.NewOsFs(), dir)), nil
}

// clone resolves options.Ref against the remote: an empty ref clones the default
// branch; otherwise the ref is tried as a branch, then a tag, then a commit.
func (f gitFactory) clone(ctx context.Context, dir string, options extension.Options) error {
	base, err := f.cloneOptions(options)
	if err != nil {
		return err
	}

	if options.Ref == "" {
		return f.run(ctx, dir, base, plumbing.ReferenceName(""))
	}

	for _, ref := range []plumbing.ReferenceName{
		plumbing.NewBranchReferenceName(options.Ref),
		plumbing.NewTagReferenceName(options.Ref),
	} {
		if err := f.run(ctx, dir, base, ref); err == nil {
			return nil
		}
	}

	return f.checkoutCommit(ctx, dir, base, options.Ref)
}

// cloneOptions assembles the shared clone configuration: credentials, a shallow
// depth, and the TLS posture the git transport honors.
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

// run executes a single shallow clone attempt, pinned to ref when one is given.
func (f gitFactory) run(ctx context.Context, dir string, base gogit.CloneOptions, ref plumbing.ReferenceName) error {
	opts := base
	if ref != "" {
		opts.ReferenceName = ref
		opts.SingleBranch = true
	}

	if _, err := gogit.PlainCloneContext(ctx, dir, false, &opts); err != nil {
		f.reset(dir)
		return fmt.Errorf("%w: clone %q: %w", api.ErrRuntime, opts.URL, err)
	}

	return nil
}

// checkoutCommit performs a full clone and checks out an arbitrary commit — the
// only ref kind a shallow single-branch clone cannot resolve.
func (f gitFactory) checkoutCommit(ctx context.Context, dir string, base gogit.CloneOptions, ref string) error {
	opts := base
	opts.Depth = 0
	opts.ReferenceName = ""
	opts.SingleBranch = false

	repo, err := gogit.PlainCloneContext(ctx, dir, false, &opts)
	if err != nil {
		f.reset(dir)
		return fmt.Errorf("%w: clone %q: %w", api.ErrRuntime, opts.URL, err)
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return fmt.Errorf("%w: resolve ref %q: %w", api.ErrRuntime, ref, err)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("%w: resolve ref %q: %w", api.ErrRuntime, ref, err)
	}

	if err := tree.Checkout(&gogit.CheckoutOptions{Hash: *hash}); err != nil {
		return fmt.Errorf("%w: resolve ref %q: %w", api.ErrRuntime, ref, err)
	}

	return nil
}

// reset empties dir so a subsequent clone attempt starts from a clean directory.
func (gitFactory) reset(dir string) {
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, 0o700)
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

// cleanupWhenDone schedules removal of dir once ctx is done.
func (f gitFactory) cleanupWhenDone(ctx context.Context, dir string) {
	f.pool.Submit(func() {
		<-ctx.Done()
		_ = os.RemoveAll(dir)
	})
}
