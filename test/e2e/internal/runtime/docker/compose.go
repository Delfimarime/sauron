package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/testcontainers/testcontainers-go/log"

	"github.com/testcontainers/testcontainers-go/modules/compose"
)

type dockerRuntime struct {
	bin       string
	directory string
	logger    log.Logger
	specs     []ContainerSpec
	stack     compose.ComposeStack
}

func New(bin, directory string, opts ...func(*Options)) (*dockerRuntime, error) {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	specs := make([]ContainerSpec, 0, len(options.specs)+1)

	specs = append(specs, ContainerSpec{
		Service:    "main",
		Image:      "alpine:",
		Entrypoint: "sleep infinity",
		Env: map[string]string{
			"SAURON_HOME": "/root/.sauron/",
		},
		Mount: append(options.mount, FileSpec{
			SourceFile: bin,
			Path:       "/opt/bin/sauron",
		}),
	})

	for index, each := range options.specs {
		if index != 0 && each.Service == "main" {
			return nil, fmt.Errorf("Cannot have multiple containers named 'main'")
		}
		specs = append(specs, each)
	}

	return &dockerRuntime{bin: bin, directory: directory, specs: specs}, nil
}

func (c *dockerRuntime) IsReadOnly() bool { return false }

func (c *dockerRuntime) Start(ctx context.Context) error {
	doc, err := GenerateDockerComposeFile(c.specs)
	if err != nil {
		return err
	}
	path := filepath.Join(c.directory, "docker-compose.yml")
	if err := os.WriteFile(path, doc, 0o600); err != nil {
		return err
	}
	stack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles(path), compose.WithLogger(c.logger),
	)
	if err != nil {
		return err
	}
	c.stack = stack
	return stack.Up(ctx)
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

func (c *dockerRuntime) Execute(ctx context.Context, args ...string) (int, string, error) {
	if len(args) == 0 {
		return -1, "", fmt.Errorf("arguments are mandatory and cannot be empty")
	}
	if args[0] == "sauron" {
		return c.Execute(ctx, append([]string{"/opt/bin/sauron"}, args[1:]...)...)
	}
	container, err := c.stack.ServiceContainer(ctx, "main")
	if err != nil {
		return -1, "", fmt.Errorf("an unexpected error occured while trying to access container 'main'")
	}
	exitCode, r, err := container.Exec(ctx, args)
	if err != nil {
		return -1, "", fmt.Errorf(`an unexpected error occured while trying executing command "%s" in container main `, strings.Join(args, ""))
	}
	binary, err := io.ReadAll(r)
	if err != nil {
		return -1, "", fmt.Errorf("an unexpected error occured while reading stdOut")
	}
	return exitCode, string(binary), nil
}
