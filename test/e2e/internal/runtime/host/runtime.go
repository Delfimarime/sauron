package host

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

type hostRuntime struct {
	bin string
}

func New(bin string) *hostRuntime { return &hostRuntime{bin: bin} }

func (h *hostRuntime) IsReadOnly() bool { return true }

func (h *hostRuntime) Start(context.Context) error { return nil }

func (h *hostRuntime) Stop(context.Context) error { return nil }

func (c *hostRuntime) CopyTo(ctx context.Context, locationURI string, content []byte) error {
	return nil
}

func (h *hostRuntime) Execute(ctx context.Context, command ...string) (int, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, h.bin, command...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		var prob *exec.ExitError
		if !errors.As(err, &prob) {
			return 0, "", err
		}
		return prob.ExitCode(), stderr.String(), nil
	}

	return 0, stdout.String(), nil
}
