package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/httpregistry"
)

// mainService is the compose service that hosts the binary under test;
// sauronPath is where the binary is mounted inside it.
const (
	mainService = "main"
	sauronHome  = "/root/.sauron/"
	sauronPath  = "/opt/bin/sauron"

	// hostDockerInternal is the DNS name the containerized binary uses to reach the
	// in-process http registry fixture running on the host; it resolves through the
	// host-gateway extra_hosts entry wired onto "main".
	hostDockerInternal = "host.docker.internal"
)

type dockerRuntime struct {
	bin        string
	directory  string
	logger     log.Logger
	specs      []ContainerSpec
	folders    map[string]*folderSource
	webservers map[string]*httpregistry.Source
	gits       map[string]*gitSource
	stack      compose.ComposeStack
}

// New builds a compose-backed runtime. It always prepends the "main" service: an
// alpine container kept alive with the binary under test bind-mounted at
// sauronPath. Extra services come from WithContainer; extra host files mounted
// into "main" come from WithFile.
func New(bin, directory string, opts ...func(*Options)) (*dockerRuntime, error) {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	specs := make([]ContainerSpec, 0, len(options.specs)+1)
	specs = append(specs, ContainerSpec{
		Service:    mainService,
		Image:      "alpine:3",
		Entrypoint: "tail -f /dev/null",
		Env: map[string]string{
			"SAURON_HOME": sauronHome,
			// Pin HOME so the git ssh transport resolves its default known_hosts to
			// the pinned host-key entry mounted at /root/.ssh/known_hosts (docker exec
			// does not set HOME), keeping strict host-key verification working.
			"HOME": "/root",
		},
		Mount: append(options.mount, FileSpec{
			SourceFile: bin,
			Path:       sauronPath,
		}),
	})

	for _, each := range options.specs {
		if each.Service == mainService {
			return nil, fmt.Errorf("docker: %q is reserved for the binary under test", mainService)
		}
		specs = append(specs, each)
	}

	return &dockerRuntime{
		bin:        bin,
		directory:  directory,
		logger:     options.logger,
		specs:      specs,
		folders:    map[string]*folderSource{},
		webservers: map[string]*httpregistry.Source{},
		gits:       map[string]*gitSource{},
	}, nil
}

func (c *dockerRuntime) IsReadOnly() bool { return false }

func (c *dockerRuntime) Start(ctx context.Context) error {
	if err := os.MkdirAll(c.directory, 0o755); err != nil {
		return fmt.Errorf("create compose directory: %w", err)
	}

	declared, err := buildSpecs(c.specs, c.folders, c.webservers, c.gits)
	if err != nil {
		return err
	}

	specs, err := materializeContent(os.WriteFile, c.directory, declared)
	if err != nil {
		return err
	}

	doc, err := GenerateDockerComposeFile(specs)
	if err != nil {
		return err
	}

	path := filepath.Join(c.directory, "docker-compose.yml")
	if err := os.WriteFile(path, doc, 0o600); err != nil {
		return fmt.Errorf("write compose file: %w", err)
	}

	stack, err := c.newStack(path)
	if err != nil {
		return err
	}
	// The git sidecar's entrypoint seeds the repo before `exec sshd`, so port 22
	// is not open the instant the container starts; wait until sshd announces it
	// is listening, otherwise an early clone races it and is refused.
	for _, alias := range sortedKeys(c.gits) {
		stack = stack.WaitForService(gitService(alias), wait.ForLog("Server listening on"))
	}
	// Webserver sources need no readiness wait: the http registry fixture runs in the
	// test process and its listener backlog accepts connections the instant it binds
	// (it binds lazily when #{.webserver.*.url} is first resolved, before the binary's
	// validation probe runs).
	c.stack = stack

	return stack.Up(ctx)
}

// newStack builds the compose stack, attaching the logger only when one was
// supplied (compose.WithLogger(nil) would install a nil logger).
func (c *dockerRuntime) newStack(path string) (compose.ComposeStack, error) {
	if c.logger != nil {
		return compose.NewDockerComposeWith(compose.WithStackFiles(path), compose.WithLogger(c.logger))
	}
	return compose.NewDockerComposeWith(compose.WithStackFiles(path))
}

func (c *dockerRuntime) Stop(ctx context.Context) error {
	for _, alias := range sortedKeys(c.webservers) {
		_ = c.webservers[alias].Server().Stop(ctx)
	}
	var err error
	if c.stack != nil {
		err = c.stack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveVolumes(true))
	}
	if c.directory != "" {
		_ = os.RemoveAll(c.directory)
		c.directory = ""
	}
	return err
}

func (c *dockerRuntime) CopyTo(ctx context.Context, locationURI string, content []byte) error {
	if c.stack == nil {
		// Seed before Start: defer the bytes as an inline-content mount on "main",
		// materialized when the stack comes up. The bind target must be the absolute
		// in-container path (a relative compose target is invalid), and the mount
		// must be written back into c.specs by index — ranging by value would discard
		// it on the loop variable's copy.
		target := containerPath(locationURI)
		for i := range c.specs {
			if c.specs[i].Service != mainService {
				continue
			}
			c.specs[i].Mount = append(c.specs[i].Mount, FileSpec{
				Content: content,
				Path:    target,
			})
		}
		return nil
	}
	container, err := c.stack.ServiceContainer(ctx, mainService)
	if err != nil {
		return fmt.Errorf("access container %q: %w", mainService, err)
	}
	if strings.HasPrefix(locationURI, "~/") {
		locationURI = "/root/" + locationURI[2:]
	}
	err = container.CopyToContainer(ctx, content, locationURI, 0700)
	if err != nil {
		return fmt.Errorf("an unexpected error occured while attempting to copy file into container.\ncaused by:%w", err)
	}
	return nil
}

// serviceExec runs a command inside a named compose service, returning the exit
// code and the relevant output stream. The git source injects it to read its
// revision from its own sidecar.
type serviceExec func(ctx context.Context, service string, args ...string) (int, string, error)

// Execute runs args inside the "main" container. "sauron" as arg0 is rewritten to
// the mounted binary path. It mirrors the host runtime's contract: the returned
// string is stdout on success and stderr on a non-zero exit.
func (c *dockerRuntime) Execute(ctx context.Context, args ...string) (int, string, error) {
	if len(args) == 0 {
		return -1, "", fmt.Errorf("docker: command arguments are mandatory")
	}
	if args[0] == "sauron" {
		args = append([]string{sauronPath}, args[1:]...)
	}
	return c.execIn(ctx, mainService, args...)
}

// execIn runs args inside the named service's container, splitting Docker's raw
// multiplexed stream into stdout/stderr so it can return the relevant one per the
// runtime contract.
func (c *dockerRuntime) execIn(ctx context.Context, service string, args ...string) (int, string, error) {
	if len(args) == 0 {
		return -1, "", fmt.Errorf("docker: command arguments are mandatory")
	}

	container, err := c.stack.ServiceContainer(ctx, service)
	if err != nil {
		return -1, "", fmt.Errorf("access container %q: %w", service, err)
	}

	exitCode, reader, err := container.Exec(ctx, args)
	if err != nil {
		return -1, "", fmt.Errorf("exec %q in %q: %w", strings.Join(args, " "), service, err)
	}

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return -1, "", fmt.Errorf("demux output of %q in %q: %w", strings.Join(args, " "), service, err)
	}

	if exitCode != 0 {
		return exitCode, stderr.String(), nil
	}
	return exitCode, stdout.String(), nil
}

// ReadFile cats a file out of the "main" container. A relative path is resolved
// against $SAURON_HOME, a "~/" path against /root, and an absolute path is read
// as-is.
func (c *dockerRuntime) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if c.stack == nil {
		return nil, fmt.Errorf("docker: runtime not started; nothing to read at %q", path)
	}
	code, out, err := c.Execute(ctx, "cat", containerPath(path))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	if code != 0 {
		return nil, fmt.Errorf("read %q: cat exited %d: %s", path, code, out)
	}
	return []byte(out), nil
}

// containerPath resolves a home-relative path to its absolute in-container location.
func containerPath(path string) string {
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		return "/root/" + rest
	}
	if strings.HasPrefix(path, "/") {
		return path
	}
	return strings.TrimRight(sauronHome, "/") + "/" + path
}

// Folder declares a content directory served from inside "main". The source is
// created once per alias so repeated Expose calls accumulate.
func (c *dockerRuntime) Folder(alias string) runtime.Source {
	src, ok := c.folders[alias]
	if !ok {
		src = &folderSource{alias: alias}
		c.folders[alias] = src
	}
	return src
}

// Webserver declares an http registry source served by the in-process fixture in the
// test process; the containerized binary reaches it at host.docker.internal. The
// source is created once per alias so repeated Expose calls accumulate.
func (c *dockerRuntime) Webserver(alias string) runtime.Source {
	src, ok := c.webservers[alias]
	if !ok {
		src = httpregistry.NewSource(httpregistry.New(), hostDockerInternal)
		c.webservers[alias] = src
	}
	return src
}

// Git declares an ssh git-remote source served by an sshd sidecar that seeds a bare
// repo from the exposed content. The source is created once per alias (with its own
// generated key material) so repeated Expose calls accumulate.
func (c *dockerRuntime) Git(alias string) runtime.Source {
	src, ok := c.gits[alias]
	if !ok {
		keys, err := newSSHKeyPair()
		if err != nil {
			return runtime.NewErroringSource(fmt.Errorf("docker: generate ssh keys for git source %q: %w", alias, err))
		}
		src = &gitSource{alias: alias, keys: keys, exec: c.execIn}
		c.gits[alias] = src
	}
	return src
}
