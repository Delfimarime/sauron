package gherkin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// setProviderController owns the set-provider assertions no generic step covers:
// the Provider recorded in settings.yaml (graybox — decoded into pkg/sauron/types)
// and the verbatim confirmation lines the command prints. State is read through the
// runtime; the printed line is read through the commandController that captured it,
// so the command result stays consumed in one place rather than re-run here.
type setProviderController struct {
	rt       runtime.Runtime
	commands *commandController
}

func (c *setProviderController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the provider is set to (\S+)$`, c.providerIsSetTo)
	sc.Step(`^the output reports provider (\S+) was set$`, c.reportsProviderSet)
	sc.Step(`^the output reports provider (\S+) was already set$`, c.reportsProviderAlreadySet)
}

// providerIsSetTo asserts settings.yaml records exactly one Provider with the given
// name (FR-001/FR-003: the chosen provider is recorded, or left unchanged).
func (c *setProviderController) providerIsSetTo(ctx context.Context, name string) error {
	provider, err := c.oneProvider(ctx)
	if err != nil {
		return err
	}
	return assertExpected("provider name", name, provider.Metadata.Name)
}

// reportsProviderSet asserts the exact change confirmation. Nothing is installed
// until install (0007) exists, so the migration count is always zero and the
// summary carries no count — the line is just the confirmation.
func (c *setProviderController) reportsProviderSet(_ context.Context, name string) error {
	return c.outputIs(fmt.Sprintf("provider set to %q", name))
}

// reportsProviderAlreadySet asserts the exact FR-003 no-change notice.
func (c *setProviderController) reportsProviderAlreadySet(_ context.Context, name string) error {
	return c.outputIs(fmt.Sprintf("provider already set to %q", name))
}

// outputIs asserts the command's printed output equals want once surrounding
// whitespace is trimmed, so a stray migration count cannot slip past a substring
// check.
func (c *setProviderController) outputIs(want string) error {
	out, err := c.commands.lastOutput()
	if err != nil {
		return err
	}
	return assertExpected("command output", want, strings.TrimSpace(out))
}

// oneProvider reads the single recorded provider, erroring unless exactly one is
// present.
func (c *setProviderController) oneProvider(ctx context.Context) (types.Provider, error) {
	data, err := c.rt.ReadFile(ctx, settingsFile)
	if err != nil {
		return types.Provider{}, err
	}
	providers, err := decodeProviders(data)
	if err != nil {
		return types.Provider{}, err
	}
	return oneProvider(providers)
}

// decodeProviders decodes a multi-document YAML stream and keeps the Provider
// documents (skipping the Registry that shares settings.yaml). Pure (bytes in,
// values out) so it is unit-tested without the fs.
func decodeProviders(data []byte) ([]types.Provider, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var out []types.Provider
	for {
		var doc types.Provider
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", settingsFile, err)
		}
		if doc.Kind != types.KindProvider {
			continue // skip empty documents and the Registry
		}
		out = append(out, doc)
	}
	return out, nil
}

// oneProvider returns the single provider, erroring unless exactly one is present.
func oneProvider(providers []types.Provider) (types.Provider, error) {
	switch len(providers) {
	case 1:
		return providers[0], nil
	case 0:
		return types.Provider{}, fmt.Errorf("no provider is configured")
	default:
		return types.Provider{}, fmt.Errorf("expected exactly one provider, found %d", len(providers))
	}
}
