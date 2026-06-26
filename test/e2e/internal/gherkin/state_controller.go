package gherkin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

const (
	// settingsFile holds the single Registry document (alongside the Provider),
	// relative to $SAURON_HOME (the runtime resolves it per backend).
	settingsFile = "settings.yaml"
	// trackFile holds the installed Skill and Agent documents.
	trackFile = "track.yaml"
)

// stateController owns the file-based Then steps: it reads the persisted state
// through the runtime, decodes it into the public pkg/sauron/types (graybox — no
// internal/), and asserts on the result. There is a single, nameless registry, so
// the registry assertions read "the registry" rather than naming one.
type stateController struct {
	rt runtime.Runtime
}

func (c *stateController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^there is exactly one registry$`, c.exactlyOneRegistry)
	sc.Step(`^there is no registry$`, c.noRegistry)
	sc.Step(`^the registry has transport (\S+)$`, c.registryTransport)
	sc.Step(`^the registry is described by:$`, c.registryDescribedBy)
	sc.Step(`^the registry stores password as the reference (\S+)$`, c.registryPasswordRef)
	sc.Step(`^the registry has a creation timestamp$`, c.registryHasCreationTimestamp)
	sc.Step(`^the stored state does not contain (.+)$`, c.configDoesNotContain)
	sc.Step(`^the skill (\S+) is still tracked$`, c.skillStillTracked)
}

func (c *stateController) exactlyOneRegistry(ctx context.Context) error {
	return c.registryCount(ctx, 1)
}

func (c *stateController) noRegistry(ctx context.Context) error {
	return c.registryCount(ctx, 0)
}

func (c *stateController) registryCount(ctx context.Context, want int) error {
	regs, err := c.readRegistries(ctx)
	if err != nil {
		return err
	}
	return assertExpected("registry count", want, len(regs))
}

func (c *stateController) registryTransport(ctx context.Context, transport string) error {
	reg, err := c.oneRegistry(ctx)
	if err != nil {
		return err
	}
	return assertExpected("registry transport", transport, string(reg.Spec.Transport))
}

func (c *stateController) registryDescribedBy(ctx context.Context, table *godog.Table) error {
	reg, err := c.oneRegistry(ctx)
	if err != nil {
		return err
	}
	fields, err := tableFields(table)
	if err != nil {
		return err
	}
	for field, want := range fields {
		got, err := registryField(reg, field)
		if err != nil {
			return err
		}
		if err := assertExpected(field, want, got); err != nil {
			return err
		}
	}
	return nil
}

func (c *stateController) registryPasswordRef(ctx context.Context, ref string) error {
	reg, err := c.oneRegistry(ctx)
	if err != nil {
		return err
	}
	if reg.Spec.Credentials == nil {
		return fmt.Errorf("the registry has no credentials block")
	}
	return assertExpected("password reference", ref, reg.Spec.Credentials.Password)
}

// registryHasCreationTimestamp proves set stamps the audit timestamps: both are
// present, parse as RFC3339, and are equal on create. The instant itself is not
// asserted — time is the real wall clock here, so only presence and format are
// checked.
func (c *stateController) registryHasCreationTimestamp(ctx context.Context) error {
	reg, err := c.oneRegistry(ctx)
	if err != nil {
		return err
	}
	created := reg.Metadata.CreatedAt
	updated := reg.Metadata.LastUpdatedAt
	if created == "" {
		return fmt.Errorf("the registry has no createdAt")
	}
	if _, err := time.Parse(time.RFC3339, created); err != nil {
		return fmt.Errorf("registry createdAt %q is not RFC3339: %w", created, err)
	}
	if _, err := time.Parse(time.RFC3339, updated); err != nil {
		return fmt.Errorf("registry lastUpdatedAt %q is not RFC3339: %w", updated, err)
	}
	return assertExpected("audit timestamps equal on create", created, updated)
}

// configDoesNotContain proves a resolved secret is never persisted: it reads the
// raw bytes of settings.yaml under $SAURON_HOME and asserts the substring is absent
// (the stored credentials must be ${env:VAR} references, not values).
func (c *stateController) configDoesNotContain(ctx context.Context, secret string) error {
	data, err := c.rt.ReadFile(ctx, settingsFile)
	if err != nil {
		return err
	}
	if bytes.Contains(data, []byte(secret)) {
		return fmt.Errorf("stored state unexpectedly contains %q", secret)
	}
	return nil
}

// skillStillTracked proves unset preserves installed artifacts: it reads track.yaml
// and asserts the named Skill is still present (unset removes the registry but never
// the track file).
func (c *stateController) skillStillTracked(ctx context.Context, name string) error {
	data, err := c.rt.ReadFile(ctx, trackFile)
	if err != nil {
		return err
	}
	skills, err := decodeSkills(data)
	if err != nil {
		return err
	}
	for _, skill := range skills {
		if skill.Metadata.Name == name {
			return nil
		}
	}
	return fmt.Errorf("skill %q is no longer tracked (have %d skills)", name, len(skills))
}

// readRegistries reads and decodes the Registry documents in settings.yaml through
// the runtime.
func (c *stateController) readRegistries(ctx context.Context) ([]types.Registry, error) {
	data, err := c.rt.ReadFile(ctx, settingsFile)
	if err != nil {
		return nil, err
	}
	return decodeRegistries(data)
}

// oneRegistry reads the single configured registry, erroring unless exactly one is
// present.
func (c *stateController) oneRegistry(ctx context.Context) (types.Registry, error) {
	regs, err := c.readRegistries(ctx)
	if err != nil {
		return types.Registry{}, err
	}
	return oneRegistry(regs)
}

// decodeRegistries decodes a multi-document YAML stream and keeps the Registry
// documents (skipping the Provider that shares settings.yaml). Pure (bytes in,
// values out) so it is unit-tested without the fs.
func decodeRegistries(data []byte) ([]types.Registry, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var out []types.Registry
	for {
		var doc types.Registry
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", settingsFile, err)
		}
		if doc.Kind != types.KindRegistry {
			continue // skip empty documents and the Provider
		}
		out = append(out, doc)
	}
	return out, nil
}

// decodeSkills decodes a multi-document YAML stream and keeps the Skill documents
// (skipping the Agents that share track.yaml). Pure so it is unit-tested without the
// fs.
func decodeSkills(data []byte) ([]types.Artifact, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var out []types.Artifact
	for {
		var doc types.Artifact
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", trackFile, err)
		}
		if doc.Kind != types.KindSkill {
			continue
		}
		out = append(out, doc)
	}
	return out, nil
}

// oneRegistry returns the single registry, erroring unless exactly one is present.
func oneRegistry(regs []types.Registry) (types.Registry, error) {
	switch len(regs) {
	case 1:
		return regs[0], nil
	case 0:
		return types.Registry{}, fmt.Errorf("no registry is configured")
	default:
		return types.Registry{}, fmt.Errorf("expected exactly one registry, found %d", len(regs))
	}
}

// registryField reads a described-by table field off a registry by its dotted name.
func registryField(reg types.Registry, field string) (string, error) {
	switch strings.TrimSpace(field) {
	case "kind":
		return reg.Kind, nil
	case "apiVersion":
		return reg.APIVersion, nil
	case "spec.transport":
		return string(reg.Spec.Transport), nil
	case "spec.source":
		return reg.Spec.Source, nil
	case "spec.revision":
		return reg.Spec.Revision, nil
	default:
		return "", fmt.Errorf("unknown registry field %q", field)
	}
}
