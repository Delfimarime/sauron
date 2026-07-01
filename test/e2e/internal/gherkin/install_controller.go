package gherkin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// installController owns the install-skill and install-agent Gherkin steps. It
// asserts on the plan output (via the commandController) and on the track.yaml
// state (via the runtime), driving the binary end-to-end without touching any
// internal/ package. A hostsAgent step is included here so agent install scenarios
// can expose content on the default http registry fixture without a separate
// controller.
type installController struct {
	rt       runtime.Runtime
	commands *commandController
}

func newInstallController(rt runtime.Runtime, commands *commandController) *installController {
	return &installController{rt: rt, commands: commands}
}

// Init registers every install-related step against the scenario context.
func (c *installController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the http server hosts an agent named (\S+)$`, c.hostsAgent)
	sc.Step(`^the skill (\S+) is tracked with a non-empty version$`, c.skillTrackedWithVersion)
	sc.Step(`^the skill (\S+) has spec\.path (.+)$`, c.skillHasPath)
	sc.Step(`^the agent (\S+) is tracked with a non-empty version$`, c.agentTrackedWithVersion)
	sc.Step(`^the agent (\S+) has spec\.path (.+)$`, c.agentHasPath)
}

// hostsAgent exposes a minimal Agent manifest on the default http registry fixture,
// mirroring the registryHTTPController's hostsSkill step for the agent kind.
func (c *installController) hostsAgent(_ context.Context, name string) error {
	c.rt.Webserver(defaultAlias).Expose(agentResource(name))
	return nil
}

func (c *installController) skillTrackedWithVersion(ctx context.Context, name string) error {
	return c.artifactTrackedWithVersion(ctx, types.KindSkill, name)
}

func (c *installController) skillHasPath(ctx context.Context, name, path string) error {
	return c.artifactHasPath(ctx, types.KindSkill, name, path)
}

func (c *installController) agentTrackedWithVersion(ctx context.Context, name string) error {
	return c.artifactTrackedWithVersion(ctx, types.KindAgent, name)
}

func (c *installController) agentHasPath(ctx context.Context, name, path string) error {
	return c.artifactHasPath(ctx, types.KindAgent, name, path)
}

func (c *installController) artifactTrackedWithVersion(ctx context.Context, kind, name string) error {
	a, err := c.readTrackedArtifact(ctx, kind, name)
	if err != nil {
		return err
	}
	if a.Spec.Version == "" {
		return fmt.Errorf("%s %q: expected a non-empty version in track.yaml but got empty", kind, name)
	}
	return nil
}

func (c *installController) artifactHasPath(ctx context.Context, kind, name, want string) error {
	a, err := c.readTrackedArtifact(ctx, kind, name)
	if err != nil {
		return err
	}
	return assertExpected(fmt.Sprintf("%s %q spec.path", kind, name), want, a.Spec.Path)
}

// readTrackedArtifact reads track.yaml through the runtime and returns the named
// artifact of the given kind, erroring when it is absent.
func (c *installController) readTrackedArtifact(ctx context.Context, kind, name string) (types.Artifact, error) {
	data, err := c.rt.ReadFile(ctx, trackFile)
	if err != nil {
		return types.Artifact{}, err
	}
	artifacts, err := decodeArtifactsOfKind(data, kind)
	if err != nil {
		return types.Artifact{}, err
	}
	for _, a := range artifacts {
		if a.Metadata.Name == name {
			return a, nil
		}
	}
	return types.Artifact{}, fmt.Errorf(
		"%s %q not found in track.yaml (have %d of that kind)", kind, name, len(artifacts),
	)
}

// decodeArtifactsOfKind decodes a multi-document YAML stream and keeps the documents
// whose Kind matches the given kind. Pure (bytes in, values out) so it is
// unit-testable without the runtime.
func decodeArtifactsOfKind(data []byte, kind string) ([]types.Artifact, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var out []types.Artifact
	for {
		var doc types.Artifact
		if err := dec.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode %s: %w", trackFile, err)
		}
		if doc.Kind != kind {
			continue
		}
		out = append(out, doc)
	}
	return out, nil
}
