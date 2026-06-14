package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// RegisterSteps wires the world lifecycle and step definitions into a godog
// scenario context. It is the single registration entrypoint the suite calls per
// scenario, so each scenario gets a fresh world.
func RegisterSteps(sc *godog.ScenarioContext) {
	w, err := NewWorld()
	if err != nil {
		panic(err)
	}

	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		return ctx, w.Reset(ctx)
	})

	// The run step requires a leading dash so it never collides with
	// "sauron version is …".
	sc.Step(`^sauron (-.+)$`, w.runSauron)
	sc.Step(`^the output should be:$`, w.theOutputShouldBe)
	sc.Step(`^sauron version is (.+)$`, w.sauronVersionIs)
}

func (w *World) runSauron(ctx context.Context, args string) error {
	return w.Execute(ctx, strings.Fields(args)...)
}

func (w *World) theOutputShouldBe(doc *godog.DocString) error {
	t := &capturingT{}

	res := w.Last()
	if !assert.NotNil(t, res, "no command has run in this scenario") {
		return t.err()
	}

	expected, err := render(doc.Content, w)
	if !assert.NoError(t, err, "rendering the expected output template") {
		return t.err()
	}

	assert.Equal(t, normalizeBanner(expected), normalizeBanner(res.output),
		"command output did not match the expected banner")
	return t.err()
}

func (w *World) sauronVersionIs(expr string) error {
	t := &capturingT{}

	res := w.Last()
	if !assert.NotNil(t, res, "no command has run in this scenario") {
		return t.err()
	}

	expected, err := render(expr, w)
	if !assert.NoError(t, err, "rendering the expected version template") {
		return t.err()
	}

	actual, err := parseBannerVersion(res.output)
	if !assert.NoError(t, err, "parsing the version from %q", res.output) {
		return t.err()
	}

	assert.Equal(t, expected, actual, "reported version did not match")
	return t.err()
}

// normalizeBanner trims surrounding whitespace from each line and drops blank
// lines, so a docstring's indentation does not defeat a banner comparison.
func normalizeBanner(s string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

// parseBannerVersion extracts the version from the banner's first line,
// "<AppName> v<AppVersion>", returning the text after the final " v".
func parseBannerVersion(out string) (string, error) {
	first := strings.TrimSpace(out)
	if first == "" {
		return "", fmt.Errorf("empty output")
	}
	first = strings.TrimSpace(strings.SplitN(first, "\n", 2)[0])

	idx := strings.LastIndex(first, " v")
	if idx < 0 {
		return "", fmt.Errorf("first line %q is not %q", first, "<AppName> v<AppVersion>")
	}

	version := strings.TrimSpace(first[idx+2:])
	if version == "" {
		return "", fmt.Errorf("first line %q has an empty version", first)
	}
	return version, nil
}
