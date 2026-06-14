package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

type basicController struct {
	rt runtime.Runtime
}

func (b *basicController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^sauron version is (.+)$`, b.IsVersion)
	sc.Step(`^sauron home directory is (.+)$`, b.IsHomeDirectory)
}

// IsVersion runs `sauron --version` and asserts the reported version equals the
// feature's expected value.
func (b *basicController) IsVersion(ctx context.Context, expected string) error {
	out, err := b.execVersion(ctx)
	if err != nil {
		return err
	}
	reported, err := b.parseVersion(out)
	if err != nil {
		return err
	}
	return assertExpected("version", expected, reported)
}

// IsHomeDirectory runs `sauron --version` and asserts the reported home directory
// equals the feature's expected value.
func (b *basicController) IsHomeDirectory(ctx context.Context, expected string) error {
	out, err := b.execVersion(ctx)
	if err != nil {
		return err
	}
	reported, err := b.parseHomeDirectory(out)
	if err != nil {
		return err
	}
	return assertExpected("home directory", expected, reported)
}

// execVersion runs `sauron --version` fresh on each call (no caching) and returns
// its stdout.
func (b *basicController) execVersion(ctx context.Context) (string, error) {
	code, out, err := b.rt.Execute(ctx, "sauron", "--version")
	if err != nil {
		return "", fmt.Errorf("run sauron --version: %w", err)
	}
	if code != 0 {
		return "", fmt.Errorf("sauron --version exited %d: %s", code, out)
	}
	return out, nil
}

// parseVersion returns the text after " v" on the banner's first non-empty line
// ("<name> v<version>").
func (b *basicController) parseVersion(out string) (string, error) {
	for raw := range strings.Lines(out) {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		// Only the first non-empty (banner) line carries the version.
		if _, version, ok := strings.Cut(line, " v"); ok {
			return version, nil
		}
		break
	}
	return "", fmt.Errorf("no \"<name> v<version>\" line in output: %q", out)
}

// parseHomeDirectory returns the value of the banner's "Home: " line.
func (b *basicController) parseHomeDirectory(out string) (string, error) {
	for raw := range strings.Lines(out) {
		if home, ok := strings.CutPrefix(strings.TrimSpace(raw), "Home: "); ok {
			return home, nil
		}
	}
	return "", fmt.Errorf("no \"Home: <dir>\" line in output: %q", out)
}
