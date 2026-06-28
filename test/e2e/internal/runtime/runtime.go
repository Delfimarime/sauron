package runtime

import "context"

// Runtime is where the command-under-test executes: the host OS, or a sandbox.
// It owns execution, so in the container/compose cases the binary runs inside the
// sandbox and reaches dependencies by service name on the internal network — which
// is why no Endpoint/Env plumbing crosses the boundary.
//
// Runtime is one wide interface and the per-scenario shared-state owner: it holds
// the provisioned sources (declared via Folder/Webserver/Git) and their addresses,
// so controllers share a single handle and there is no separate "world". A
// capability a backend cannot satisfy returns an error from the relevant Source
// accessor (there is no Pod sub-interface and no rt.(Pod) type-assert — a gap is an
// honest error, not a structural check).
type Runtime interface {
	// IsReadOnly reports whether the runtime must not be mutated by a scenario.
	// The host OS is read-only (true); ephemeral sandboxes are not (false).
	IsReadOnly() bool
	// Execute runs the binary under test with command as its args, returning the
	// exit code, the relevant output stream (stdout on success, stderr on a
	// non-zero exit), and an error ONLY for harness-level failures (the process or
	// container exec could not run). A non-zero exit is not an error.
	Execute(context.Context, ...string) (int, string, error)

	// ReadFile reads a file produced by the binary under test. A relative path is
	// resolved against the runtime's $SAURON_HOME (the per-scenario temp dir on the
	// host, the in-container home under docker); an absolute path is read as-is.
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// CopyTo writes content into the runtime at path (also step-facing for inline
	// docstring content). A relative path is resolved against the runtime home.
	CopyTo(context.Context, string, []byte) error

	// Folder declares a local-directory source by alias; Source.Path() yields the
	// directory the binary can read. Supported by every backend.
	Folder(alias string) Source
	// Webserver declares an http(s) source by alias; Source.URL() yields the
	// endpoint. Unsupported on the host backend (the accessor's Source errors).
	Webserver(alias string) Source
	// Git declares an ssh git-remote source by alias; Source.URL() yields the
	// remote. Served by an sshd sidecar under docker; the host backend errors.
	Git(alias string) Source

	Start(context.Context) error

	Stop(context.Context) error // tear everything down
}

// Source is a declared, resource-loaded exposure of provider content. Given steps
// only accumulate (Expose); accessing an attribute (Path/URL) forces the lazy
// Start and returns the live address. A capability gap surfaces as an error from
// Path/URL.
type Source interface {
	// Expose declares what the source serves: content files and, for a webserver,
	// optional basic-auth credentials (never a port). Accumulates only.
	Expose(resources ...Resource)
	// Path returns the local directory the source is served from (Folder sources).
	Path(ctx context.Context) (string, error)
	// URL returns the network address the source is reachable at (Webserver/Git
	// sources).
	URL(ctx context.Context) (string, error)
	// SSHKey returns the in-runtime path of the private key the binary must present
	// to authenticate against the source (Git sources over ssh). A source that needs
	// no key returns an error.
	SSHKey(ctx context.Context) (string, error)
	// Revision returns the commit the source's content resolves to (Git sources):
	// the HEAD a scenario can pin to exercise commit-addressed resolution. A source
	// that has no revision returns an error.
	Revision(ctx context.Context) (string, error)
}

// Resource customizes a Source. A file resource (Path/Content set) adds content to
// the served set; an auth resource (Username/Password set) configures basic auth on
// a webserver source. The two never carry a port.
type Resource struct {
	// Path is the file's location within the content set, e.g. "skills/go/skill.yaml".
	Path string
	// Content is the file's bytes.
	Content []byte
	// Username and Password configure basic auth on a webserver source; both empty
	// means the resource is a content file, not an auth declaration.
	Username string
	Password string
}

// IsAuth reports whether the resource declares basic-auth credentials rather than a
// content file.
func (r Resource) IsAuth() bool { return r.Username != "" || r.Password != "" }

type Factory interface {
	New(binaryURI, directoryURI string) (Runtime, error)
}
