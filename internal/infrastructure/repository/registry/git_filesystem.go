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

// gitFactory opens a git-backed source by shallow-cloning a repository.
type gitFactory struct {
	pool pond.Pool
}

// newGitFactory builds a gitFactory that reclaims clones through the given pool.
func newGitFactory(pool pond.Pool) *gitFactory {
	return &gitFactory{pool: pool}
}

// Validate accepts every option the git transport understands. The configuration
// is verified against the remote only when the source is opened.
func (gitFactory) Validate(_ ...extension.Option) error {
	return nil
}

// Open shallow-clones the repository at the URI into a temporary directory bound
// to ctx's lifetime, checks out the requested reference, and returns a read-only
// view rooted at the clone.
func (f gitFactory) Open(ctx context.Context, opts ...extension.Option) (source.FileSystem, error) {
	options := api.Resolve(opts)

	dir, err := os.MkdirTemp("", "sauron-registry-git-")
	if err != nil {
		return nil, fmt.Errorf("%w: stage clone: %w", api.ErrRuntime, err)
	}
	f.cleanupWhenDone(ctx, dir)

	if err := f.clone(ctx, dir, options); err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}

	return api.NewDirectory(afero.NewBasePathFs(afero.NewOsFs(), dir)), nil
}

// clone performs the shallow clone, retrying an ambiguous reference as a tag
// when it is not a branch.
func (f gitFactory) clone(ctx context.Context, dir string, options extension.Options) error {
	auth, err := f.authFrom(options)
	if err != nil {
		return err
	}

	if options.Ref == "" {
		return f.run(ctx, dir, options.URI, auth, plumbing.ReferenceName(""))
	}

	branchErr := f.run(ctx, dir, options.URI, auth, plumbing.NewBranchReferenceName(options.Ref))
	if branchErr == nil {
		return nil
	}

	if tagErr := f.run(ctx, dir, options.URI, auth, plumbing.NewTagReferenceName(options.Ref)); tagErr == nil {
		return nil
	}

	return branchErr
}

// run executes a single shallow clone attempt against the given reference.
func (f gitFactory) run(ctx context.Context, dir, uri string, auth transport.AuthMethod, ref plumbing.ReferenceName) error {
	opts := &gogit.CloneOptions{
		URL:   uri,
		Auth:  auth,
		Depth: 1,
	}
	if ref != "" {
		opts.ReferenceName = ref
		opts.SingleBranch = true
	}

	if _, err := gogit.PlainCloneContext(ctx, dir, false, opts); err != nil {
		_ = os.RemoveAll(dir)
		_ = os.MkdirAll(dir, 0o700)
		return fmt.Errorf("%w: clone %q: %w", api.ErrRuntime, uri, err)
	}

	return nil
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
