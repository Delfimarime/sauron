package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// RegisterSteps wires the scenario world's lifecycle and the step definitions
// into a godog scenario context. It is the single registration entrypoint the
// suite calls per scenario.
//
// The smoke scenario asserts on the version banner, which is plain text with no
// pkg/ type to model it, so no pkg/ import appears here. Step definitions that
// decode structured command output (registry/provider/backend listings) into
// the public pkg/ types are added as feature coverage lands; they import pkg/
// and never internal/, the rule the module's depguard config enforces.
func RegisterSteps(sc *godog.ScenarioContext) {
	w, err := NewWorld()
	if err != nil {
		panic(err)
	}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.Reset()
		return ctx, nil
	})

	sc.Step(`^I run the binary with "([^"]*)"$`, w.iRunTheBinaryWith)
	sc.Step(`^the command exits successfully$`, w.theCommandExitsSuccessfully)
	sc.Step(`^the output is a version banner$`, w.theOutputIsAVersionBanner)
}

// versionBanner is the parsed shape of the --version output. It is local to the
// harness: the banner is fixed by the architecture contract's Root command
// section, not by a pkg/ port, so it has no public type.
type versionBanner struct {
	AppName    string
	AppVersion string
	Hash       string
	Home       string
}

func (w *World) iRunTheBinaryWith(ctx context.Context, args string) error {
	return w.Run(ctx, strings.Fields(args)...)
}

func (w *World) theCommandExitsSuccessfully() error {
	t := &capturingT{}

	res := w.Last()
	if !assert.NotNil(t, res, "no command has run in this scenario") {
		return t.err()
	}
	assert.Equal(t, 0, res.ExitCode, "expected a zero exit code; stderr: %s", res.Stderr)

	return t.err()
}

func (w *World) theOutputIsAVersionBanner() error {
	t := &capturingT{}

	res := w.Last()
	if !assert.NotNil(t, res, "no command has run in this scenario") {
		return t.err()
	}

	banner, err := parseVersionBanner(res.Stdout)
	if !assert.NoError(t, err, "stdout was not a version banner: %q", res.Stdout) {
		return t.err()
	}

	assert.NotEmpty(t, banner.AppName, "banner is missing the app name")
	assert.NotEmpty(t, banner.AppVersion, "banner is missing the version")
	assert.NotEmpty(t, banner.Home, "banner is missing the resolved home")

	return t.err()
}

// parseVersionBanner decodes the three-line banner the root command emits:
//
//	<AppName> v<AppVersion>
//	Hash <AppHash>
//	Home: <home>
func parseVersionBanner(out string) (versionBanner, error) {
	lines := splitNonEmptyLines(out)
	if len(lines) < 3 {
		return versionBanner{}, fmt.Errorf("expected at least 3 banner lines, got %d", len(lines))
	}

	name, version, err := parseNameVersion(lines[0])
	if err != nil {
		return versionBanner{}, err
	}

	hash, ok := strings.CutPrefix(lines[1], "Hash ")
	if !ok {
		return versionBanner{}, fmt.Errorf("second line %q has no %q prefix", lines[1], "Hash ")
	}

	home, ok := strings.CutPrefix(lines[2], "Home: ")
	if !ok {
		return versionBanner{}, fmt.Errorf("third line %q has no %q prefix", lines[2], "Home: ")
	}

	return versionBanner{
		AppName:    name,
		AppVersion: version,
		Hash:       strings.TrimSpace(hash),
		Home:       strings.TrimSpace(home),
	}, nil
}

// parseNameVersion splits the banner's first line, "<AppName> v<AppVersion>",
// into its name and version.
func parseNameVersion(line string) (string, string, error) {
	idx := strings.LastIndex(line, " v")
	if idx < 0 {
		return "", "", fmt.Errorf("first line %q is not %q", line, "<AppName> v<AppVersion>")
	}

	name := strings.TrimSpace(line[:idx])
	version := strings.TrimSpace(line[idx+2:])
	if name == "" || version == "" {
		return "", "", fmt.Errorf("first line %q has an empty name or version", line)
	}

	return name, version, nil
}

func splitNonEmptyLines(s string) []string {
	out := make([]string, 0, 3)
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			out = append(out, strings.TrimRight(line, "\r"))
		}
	}
	return out
}
