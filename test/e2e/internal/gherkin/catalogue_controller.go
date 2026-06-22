package gherkin

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// fileDirective introduces one manifest in an "offers" doc-string. Every artifact
// fixture is authored explicitly under such a directive, so the seeded layout —
// including the exact filename and extension that fixes the catalogue name — is
// designed by the feature, never synthesized in controller code.
const fileDirective = "# file:"

// catalogueController owns the catalogue fixture Given (seeding a filesystem
// registry's .skills/.agents/.personas) and the catalogue Then assertions. It reads
// the recorded output of the last command from the commandController rather than
// re-running anything, so the rendered rows and paging line are asserted exactly as
// the user saw them. Seed logic lives in pure helpers so it is unit-tested without a
// process or real fs.
type catalogueController struct {
	rt      runtime.Runtime
	command *commandController
}

func (c *catalogueController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the registry (.+) offers the following (skills|agents|personas):$`, c.offers)
	sc.Step(`^the catalogue lists (.+)$`, c.catalogueLists)
	sc.Step(`^the paging line reads (.+)$`, c.pagingLineReads)
}

// offers seeds a filesystem registry that exposes the doc-string's manifests under
// the kind's source root, then records the registry in registries.yaml pointing at
// the materialized folder. Repeated calls for the same registry accumulate onto one
// folder, so a scenario can offer skills, agents, and personas to the same source.
func (c *catalogueController) offers(ctx context.Context, registry, kind string, body *godog.DocString) error {
	resources, err := parseManifests(body.Content)
	if err != nil {
		return err
	}
	source := c.rt.Folder(registry)
	source.Expose(resources...)
	path, err := source.Path(ctx)
	if err != nil {
		return err
	}
	stream, err := filesystemRegistryStream(registry, path)
	if err != nil {
		return err
	}
	return c.rt.CopyTo(ctx, registriesFile, stream)
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
// set (e.g. ".skills/go-style.yaml"). Pure (doc-string in, resources out) so it is
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

// filesystemRegistryStream renders a one-document registries.yaml for a schema-valid
// filesystem Registry named name whose source uri is path. Pure (name+path in, bytes
// out) so it is unit-tested without the runtime.
func filesystemRegistryStream(name, path string) ([]byte, error) {
	reg := types.Registry{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
	}
	reg.Metadata.Name = name
	reg.Spec.Transport = types.TransportFilesystem
	reg.Spec.URI = path

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(reg); err != nil {
		return nil, fmt.Errorf("encode registry %q: %w", name, err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
