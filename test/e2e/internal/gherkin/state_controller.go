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

// registriesFile is the state document the registry assertions read,
// relative to $SAURON_HOME (the runtime resolves it per backend).
const registriesFile = "registries.yaml"

// stateController owns the file-based Then steps: it reads the persisted
// state through the runtime, decodes it into the public pkg/sauron/types
// (graybox — no internal/), and asserts on the result.
type stateController struct {
	rt runtime.Runtime
}

func (c *stateController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^there is exactly one registry$`, c.exactlyOneRegistry)
	sc.Step(`^there are (\d+) registries$`, c.registryCount)
	sc.Step(`^a registry named (\S+) exists$`, c.registryExists)
	sc.Step(`^the registry (\S+) has transport (\S+)$`, c.registryTransport)
	sc.Step(`^the registry (\S+) has label (\S+) with value (\S+)$`, c.registryLabel)
	sc.Step(`^the registry (\S+) is described by:$`, c.registryDescribedBy)
	sc.Step(`^the registry (\S+) stores password as the reference (\S+)$`, c.registryPasswordRef)
	sc.Step(`^the registry (\S+) has a creation timestamp$`, c.registryHasCreationTimestamp)
	sc.Step(`^the stored state does not contain (.+)$`, c.configDoesNotContain)
}

func (c *stateController) exactlyOneRegistry(ctx context.Context) error {
	return c.registryCount(ctx, 1)
}

func (c *stateController) registryCount(ctx context.Context, want int) error {
	regs, err := c.readRegistries(ctx)
	if err != nil {
		return err
	}
	return assertExpected("registry count", want, len(regs))
}

func (c *stateController) registryExists(ctx context.Context, name string) error {
	_, err := c.findRegistry(ctx, name)
	return err
}

func (c *stateController) registryTransport(ctx context.Context, name, transport string) error {
	reg, err := c.findRegistry(ctx, name)
	if err != nil {
		return err
	}
	return assertExpected("transport of "+name, transport, string(reg.Spec.Transport))
}

func (c *stateController) registryLabel(ctx context.Context, name, key, value string) error {
	reg, err := c.findRegistry(ctx, name)
	if err != nil {
		return err
	}
	got, ok := reg.Metadata.Labels[key]
	if !ok {
		return fmt.Errorf("registry %q has no label %q", name, key)
	}
	return assertExpected("label "+key+" of "+name, value, got)
}

func (c *stateController) registryDescribedBy(ctx context.Context, name string, table *godog.Table) error {
	reg, err := c.findRegistry(ctx, name)
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
		if err := assertExpected(field+" of "+name, want, got); err != nil {
			return err
		}
	}
	return nil
}

func (c *stateController) registryPasswordRef(ctx context.Context, name, ref string) error {
	reg, err := c.findRegistry(ctx, name)
	if err != nil {
		return err
	}
	if reg.Spec.Auth == nil {
		return fmt.Errorf("registry %q has no auth block", name)
	}
	return assertExpected("password reference of "+name, ref, reg.Spec.Auth.Password)
}

// registryHasCreationTimestamp proves add stamps the audit timestamps: both are
// present, parse as RFC3339, and are equal on create. The instant itself is not
// asserted — time is the real wall clock here, so only presence and format are
// checked.
func (c *stateController) registryHasCreationTimestamp(ctx context.Context, name string) error {
	reg, err := c.findRegistry(ctx, name)
	if err != nil {
		return err
	}
	created := reg.Metadata.CreationTimestamp
	updated := reg.Metadata.LastUpdatedTimestamp
	if created == "" {
		return fmt.Errorf("registry %q has no creationTimestamp", name)
	}
	if _, err := time.Parse(time.RFC3339, created); err != nil {
		return fmt.Errorf("registry %q creationTimestamp %q is not RFC3339: %w", name, created, err)
	}
	if _, err := time.Parse(time.RFC3339, updated); err != nil {
		return fmt.Errorf("registry %q lastUpdatedTimestamp %q is not RFC3339: %w", name, updated, err)
	}
	return assertExpected("audit timestamps of "+name+" equal on create", created, updated)
}

// configDoesNotContain proves a resolved secret is never persisted: it reads the
// raw bytes of registries.yaml under $SAURON_HOME and asserts the substring is
// absent (the stored credentials must be ${env:VAR} references, not values).
func (c *stateController) configDoesNotContain(ctx context.Context, secret string) error {
	data, err := c.rt.ReadFile(ctx, registriesFile)
	if err != nil {
		return err
	}
	if bytes.Contains(data, []byte(secret)) {
		return fmt.Errorf("stored state unexpectedly contains %q", secret)
	}
	return nil
}

// readRegistries reads and decodes registries.yaml through the runtime.
func (c *stateController) readRegistries(ctx context.Context) ([]types.Registry, error) {
	data, err := c.rt.ReadFile(ctx, registriesFile)
	if err != nil {
		return nil, err
	}
	return decodeRegistries(data)
}

func (c *stateController) findRegistry(ctx context.Context, name string) (types.Registry, error) {
	regs, err := c.readRegistries(ctx)
	if err != nil {
		return types.Registry{}, err
	}
	return findRegistry(regs, name)
}

// decodeRegistries decodes a multi-document YAML stream and keeps the Registry
// documents. Pure (bytes in, values out) so it is unit-tested without the fs.
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
			return nil, fmt.Errorf("decode %s: %w", registriesFile, err)
		}
		if doc.Kind != types.KindRegistry {
			continue // skip empty documents and other kinds
		}
		out = append(out, doc)
	}
	return out, nil
}

// findRegistry returns the registry with the given metadata.name.
func findRegistry(regs []types.Registry, name string) (types.Registry, error) {
	for _, reg := range regs {
		if reg.Metadata.Name == name {
			return reg, nil
		}
	}
	return types.Registry{}, fmt.Errorf("no registry named %q (have %d registries)", name, len(regs))
}

// registryField reads a described-by table field off a registry by its dotted name.
func registryField(reg types.Registry, field string) (string, error) {
	switch strings.TrimSpace(field) {
	case "kind":
		return reg.Kind, nil
	case "apiVersion":
		return reg.APIVersion, nil
	case "name", "metadata.name":
		return reg.Metadata.Name, nil
	case "spec.transport":
		return string(reg.Spec.Transport), nil
	case "spec.uri":
		return reg.Spec.URI, nil
	case "spec.ref":
		return reg.Spec.Ref, nil
	default:
		return "", fmt.Errorf("unknown registry field %q", field)
	}
}
