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
)

// mainService is the compose service that hosts the binary under test;
// sauronPath is where the binary is mounted inside it.
const (
	mainService = "main"
	sauronHome  = "/root/.sauron/"
	sauronPath  = "/opt/bin/sauron"
)

type dockerRuntime struct {
	bin       string
	directory string
	logger    log.Logger
	specs     []ContainerSpec
	stack     compose.ComposeStack
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

	return &dockerRuntime{bin: bin, directory: directory, logger: options.logger, specs: specs}, nil
}

func (c *dockerRuntime) IsReadOnly() bool { return false }

func (c *dockerRuntime) Start(ctx context.Context) error {
	if err := os.MkdirAll(c.directory, 0o755); err != nil {
		return fmt.Errorf("create compose directory: %w", err)
	}

	specs, err := materializeContent(os.WriteFile, c.directory, c.specs)
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

// Execute runs args inside the "main" container. "sauron" as arg0 is rewritten to
// the mounted binary path. It mirrors the host runtime's contract: the returned
// string is stdout on success and stderr on a non-zero exit.
func (c *dockerRuntime) Execute(ctx context.Context, args ...string) (int, string, error) {
	if len(args) == 0 {
		return -1, "", fmt.Errorf("docker: command arguments are mandatory")
	}
	if args[0] == "sauron" {
		return c.Execute(ctx, append([]string{sauronPath}, args[1:]...)...)
	}

	container, err := c.stack.ServiceContainer(ctx, mainService)
	if err != nil {
		return -1, "", fmt.Errorf("access container %q: %w", mainService, err)
	}

	// Exec (no Multiplexed option) yields Docker's raw multiplexed stream; split
	// it into stdout/stderr so we can return the relevant one per the contract.
	exitCode, reader, err := container.Exec(ctx, args)
	if err != nil {
		return -1, "", fmt.Errorf("exec %q in %q: %w", strings.Join(args, " "), mainService, err)
	}

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return -1, "", fmt.Errorf("demux output of %q in %q: %w", strings.Join(args, " "), mainService, err)
	}

	if exitCode != 0 {
		return exitCode, stderr.String(), nil
	}
	return exitCode, stdout.String(), nil
}
