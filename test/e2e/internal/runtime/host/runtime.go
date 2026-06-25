package host

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// basicAuthEnvVar/basicAuthSecret mirror the docker fixture's binding for the
// basic-auth scenario: the binary resolves the ${env:basicAuthEnvVar} credential
// reference to basicAuthSecret at connect time. The host backend cannot serve a
// webserver source, so this only keeps the two runtimes' environments at parity.
const (
	basicAuthEnvVar = "ACME_TOKEN"
	basicAuthSecret = "s3cr3t"
)

// hostRuntime executes the binary directly on the host OS. It owns one per-scenario
// directory pinned as $SAURON_HOME: the binary writes its state there, and
// CopyTo/ReadFile/Folder all operate under it, so the suite never touches the real
// ~/.sauron. Being execution + a local folder and nothing networked keeps
// IsReadOnly() honest — Webserver and Git require the sandbox and so error here.
type hostRuntime struct {
	bin     string
	dir     string
	folders map[string]*folderSource
}

// New builds a host runtime rooted at dir (the per-scenario home). dir need not
// exist yet; Start creates it.
func New(bin, dir string) *hostRuntime {
	return &hostRuntime{bin: bin, dir: dir, folders: map[string]*folderSource{}}
}

func (h *hostRuntime) IsReadOnly() bool { return true }

// Start creates the per-scenario home directory so the binary and the fixtures
// share a known root.
func (h *hostRuntime) Start(context.Context) error {
	if h.dir == "" {
		return nil
	}
	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		return fmt.Errorf("create host home %q: %w", h.dir, err)
	}
	return nil
}

func (h *hostRuntime) Stop(context.Context) error { return nil }

// resolve turns a relative path into one under the per-scenario home; absolute and
// ~/ paths are anchored at the home root too.
func (h *hostRuntime) resolve(path string) string {
	if rest, ok := trimHome(path); ok {
		path = rest
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(h.dir, path)
}

// trimHome strips a leading "~/" so home-relative paths anchor at the per-scenario
// directory rather than the real user home.
func trimHome(path string) (string, bool) {
	if len(path) >= 2 && path[0] == '~' && path[1] == '/' {
		return path[2:], true
	}
	return path, false
}

func (h *hostRuntime) CopyTo(_ context.Context, locationURI string, content []byte) error {
	target := h.resolve(locationURI)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", target, err)
	}
	if err := os.WriteFile(target, content, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", target, err)
	}
	return nil
}

func (h *hostRuntime) ReadFile(_ context.Context, path string) ([]byte, error) {
	target := h.resolve(path)
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", target, err)
	}
	return data, nil
}

// Folder returns the per-alias local-directory source, created once per alias so
// repeated Expose calls accumulate onto the same source.
func (h *hostRuntime) Folder(alias string) runtime.Source {
	src, ok := h.folders[alias]
	if !ok {
		src = &folderSource{dir: filepath.Join(h.dir, "registry", alias)}
		h.folders[alias] = src
	}
	return src
}

// Webserver and Git require the sandbox; requesting them on the host is an error,
// not a silent skip.
func (h *hostRuntime) Webserver(string) runtime.Source {
	return runtime.NewErroringSource(errors.New("host: a webserver source is not available on @no-sandbox"))
}

func (h *hostRuntime) Git(string) runtime.Source {
	return runtime.NewErroringSource(errors.New("host: a git source is not available on @no-sandbox; use the docker runtime"))
}

func (h *hostRuntime) Execute(ctx context.Context, command ...string) (int, string, error) {
	if len(command) > 0 && command[0] == "sauron" {
		command = command[1:]
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, h.bin, command...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if h.dir != "" {
		cmd.Env = append(os.Environ(), "SAURON_HOME="+h.dir)
		// Parity with the sandbox runtime: the basic-auth fixture binds a
		// ${env:VAR} credential reference to a concrete secret on the binary's
		// environment so the binary resolves it at connect time.
		cmd.Env = append(cmd.Env, basicAuthEnvVar+"="+basicAuthSecret)
	}

	if err := cmd.Run(); err != nil {
		var prob *exec.ExitError
		if !errors.As(err, &prob) {
			return 0, "", err
		}
		return prob.ExitCode(), stderr.String(), nil
	}

	return 0, stdout.String(), nil
}

// folderSource is a host directory the binary can read. Expose accumulates content;
// Path materializes it under the directory and returns the path.
type folderSource struct {
	dir       string
	resources []runtime.Resource
}

func (s *folderSource) Expose(resources ...runtime.Resource) {
	s.resources = append(s.resources, resources...)
}

func (s *folderSource) Path(_ context.Context) (string, error) {
	for _, r := range s.resources {
		if r.IsAuth() {
			continue // a folder is not authenticated
		}
		target := filepath.Join(s.dir, r.Path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", fmt.Errorf("create directory for %q: %w", target, err)
		}
		if err := os.WriteFile(target, r.Content, 0o600); err != nil {
			return "", fmt.Errorf("write resource %q: %w", target, err)
		}
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", fmt.Errorf("create folder %q: %w", s.dir, err)
	}
	return s.dir, nil
}

func (s *folderSource) URL(context.Context) (string, error) {
	return "", fmt.Errorf("host: a folder source has no URL; use its path")
}

func (s *folderSource) SSHKey(context.Context) (string, error) {
	return "", fmt.Errorf("host: a folder source has no ssh key; use its path")
}

func (s *folderSource) Revision(context.Context) (string, error) {
	return "", fmt.Errorf("host: a folder source has no revision; use its path")
}
