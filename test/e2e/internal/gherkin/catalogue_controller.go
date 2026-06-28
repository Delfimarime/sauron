package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// fileDirective introduces one manifest in an "offers" doc-string. Every artifact
// fixture is authored explicitly under such a directive, so the seeded layout —
// including the exact filename and extension that fixes the catalogue name — is
// designed by the feature, never synthesized in controller code.
const fileDirective = "# file:"

// catalogueController owns the catalogue fixture Given (exposing the doc-string's
// skills/agents on an nginx sidecar, which a later `set registry` When configures
// black-box) and the catalogue Then assertions. It reads the recorded output of the
// last command from the commandController rather than re-running anything, so the
// rendered rows and paging line are asserted exactly as the user saw them. Parse logic
// lives in pure helpers so it is unit-tested without a process or real fs.
type catalogueController struct {
	rt      runtime.Runtime
	command *commandController
}

func (c *catalogueController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the registry offers the following (skills|agents):$`, c.offers)
	sc.Step(`^the catalogue lists (.+)$`, c.catalogueLists)
	sc.Step(`^the paging line reads (.+)$`, c.pagingLineReads)
}

// offers exposes the doc-string's manifests on the single http source (an nginx
// sidecar). A later `set the http registry from #{.webserver.default.url}` When then
// configures the registry black-box. Repeated calls accumulate onto one source, so a
// scenario can offer skills and agents to the same registry.
func (c *catalogueController) offers(_ context.Context, _ string, body *godog.DocString) error {
	resources, err := parseManifests(body.Content)
	if err != nil {
		return err
	}
	c.rt.Webserver(defaultAlias).Expose(resources...)
	return nil
}

// catalogueLists asserts a "NAME KIND" row is present in the rendered catalogue. The
// argument is the whitespace-separated pair the table renders, e.g. "go-style skill".
func (c *catalogueController) catalogueLists(_ context.Context, row string) error {
	output, err := c.command.lastOutput()
	if err != nil {
		return err
	}
	if !catalogueHasRow(output, row) {
		return fmt.Errorf("catalogue does not list %q; got: %s", row, output)
	}
	return nil
}

// pagingLineReads asserts the exact paging line is present, e.g.
// "showing 2–2 (page 2, limit 1)" or "showing 0 results (page 9, limit 20)".
func (c *catalogueController) pagingLineReads(_ context.Context, line string) error {
	output, err := c.command.lastOutput()
	if err != nil {
		return err
	}
	if !hasLine(output, line) {
		return fmt.Errorf("paging line %q not found; got: %s", line, output)
	}
	return nil
}

// parseManifests splits an "offers" doc-string into one content resource per
// "# file: <path>" directive, the path being the file's location within the content
// set (e.g. "skills/go-style.yaml"). Pure (doc-string in, resources out) so it is
// unit-tested without the runtime.
func parseManifests(body string) ([]runtime.Resource, error) {
	var (
		out     []runtime.Resource
		path    string
		content strings.Builder
		open    bool
	)
	flush := func() {
		if open {
			out = append(out, runtime.Resource{Path: path, Content: []byte(content.String())})
		}
	}
	for _, raw := range strings.Split(body, "\n") {
		if rest, ok := cutDirective(raw); ok {
			flush()
			path = rest
			content.Reset()
			open = true
			continue
		}
		if !open {
			if strings.TrimSpace(raw) == "" {
				continue // leading blank lines before the first directive
			}
			return nil, fmt.Errorf("manifest content before the first %q directive: %q", fileDirective, raw)
		}
		content.WriteString(raw)
		content.WriteString("\n")
	}
	flush()
	if len(out) == 0 {
		return nil, fmt.Errorf("no %q directive found in the offers doc-string", fileDirective)
	}
	return out, nil
}

// cutDirective reports whether line is a "# file: <path>" directive and, if so,
// returns the trimmed path.
func cutDirective(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, fileDirective) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, fileDirective)), true
}

// catalogueHasRow reports whether some output line, read as whitespace-separated
// fields, equals the expected "NAME KIND" pair. Matching on fields (not a raw
// substring) keeps column alignment from defeating the assertion. Pure so it is
// unit-tested without a process.
func catalogueHasRow(output, row string) bool {
	want := strings.Fields(row)
	for raw := range strings.SplitSeq(output, "\n") {
		if equalFields(strings.Fields(raw), want) {
			return true
		}
	}
	return false
}

// hasLine reports whether some output line, trimmed, equals the expected line. Pure
// so it is unit-tested without a process.
func hasLine(output, line string) bool {
	want := strings.TrimSpace(line)
	for raw := range strings.SplitSeq(output, "\n") {
		if strings.TrimSpace(raw) == want {
			return true
		}
	}
	return false
}

// equalFields reports whether two field slices are element-wise equal.
func equalFields(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
