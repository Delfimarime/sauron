package gherkin

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// seedColumns are the table columns the registry seed step understands; the
// username/password columns populate spec.auth, and every other column populates the
// matching spec field. name is accepted but unused for the single registry (its
// identity is spec.uri).
var seedColumns = map[string]struct{}{
	"name": {}, "transport": {}, "uri": {}, "ref": {}, "timeout": {},
	"username": {}, "password": {}, "sshKey": {},
	"creationTimestamp": {}, "lastUpdatedTimestamp": {},
}

// seedController owns the arrange Given seeds (the graybox-arrange exception): it
// writes a schema-valid public document stream that a read-only command can then
// consume. Like every controller it holds only the runtime; the seed logic lives in
// pure helpers so it is unit-tested without a process or real fs.
type seedController struct {
	rt runtime.Runtime
}

func (c *seedController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the registry is configured:$`, c.registryConfigured)
	sc.Step(`^the settings file contains:$`, c.settingsFileContains)
	sc.Step(`^a tracked skill named (\S+)$`, c.trackedSkill)
}

// registryConfigured seeds the single schema-valid Registry document into
// settings.yaml under $SAURON_HOME from a one-row table.
func (c *seedController) registryConfigured(ctx context.Context, table *godog.Table) error {
	stream, err := buildRegistryStream(table)
	if err != nil {
		return err
	}
	return c.rt.CopyTo(ctx, settingsFile, stream)
}

// settingsFileContains writes the exact bytes the scenario provides into
// settings.yaml, so the file under test — including a deliberately malformed one — is
// designed explicitly by the feature, never synthesized in code.
func (c *seedController) settingsFileContains(ctx context.Context, content *godog.DocString) error {
	return c.rt.CopyTo(ctx, settingsFile, []byte(content.Content))
}

// trackedSkill seeds one schema-valid installed Skill into track.yaml so a scenario
// can prove unset preserves installed artifacts. unset never writes track.yaml, so
// seeding it cannot mask a defect in the path under test.
func (c *seedController) trackedSkill(ctx context.Context, name string) error {
	return c.rt.CopyTo(ctx, trackFile, trackedSkillStream(name))
}

// buildRegistryStream turns a |transport|uri|…| table into a multi-document YAML
// stream of schema-valid Registry documents. Pure (table in, bytes out) so it is
// unit-tested without the runtime.
func buildRegistryStream(table *godog.Table) ([]byte, error) {
	if table == nil || len(table.Rows) < 2 {
		return nil, fmt.Errorf("the registry is configured: needs a header and at least one row")
	}
	header := table.Rows[0].Cells
	for _, cell := range header {
		if _, ok := seedColumns[cell.Value]; !ok {
			return nil, fmt.Errorf("unknown registry column %q (valid: name, transport, uri, ref, timeout, username, password, sshKey, creationTimestamp, lastUpdatedTimestamp)", cell.Value)
		}
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	for i, row := range table.Rows[1:] {
		reg, err := registryFromRow(header, row.Cells)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+1, err)
		}
		if err := enc.Encode(reg); err != nil {
			return nil, fmt.Errorf("encode registry: %w", err)
		}
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// registryFromRow maps one table row to a schema-valid Registry, stamping the
// apiVersion/kind envelope and populating the spec fields the columns carry. The
// single registry has no name; a name column, if present, is carried verbatim but is
// not required.
func registryFromRow(header, cells []*messages.PickleTableCell) (types.Registry, error) {
	if len(cells) != len(header) {
		return types.Registry{}, fmt.Errorf("expected %d cells, got %d", len(header), len(cells))
	}
	reg := types.Registry{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindRegistry},
	}
	for i, head := range header {
		value := cells[i].Value
		switch head.Value {
		case "name":
			reg.Metadata.Name = value
		case "transport":
			reg.Spec.Transport = types.Transport(value)
		case "uri":
			reg.Spec.URI = value
		case "ref":
			reg.Spec.Ref = value
		case "timeout":
			reg.Spec.Timeout = value
		case "username":
			ensureAuth(&reg).Username = value
		case "password":
			ensureAuth(&reg).Password = value
		case "sshKey":
			reg.Spec.SSHKey = value
		case "creationTimestamp":
			reg.Metadata.CreationTimestamp = value
		case "lastUpdatedTimestamp":
			reg.Metadata.LastUpdatedTimestamp = value
		}
	}
	return reg, nil
}

// ensureAuth lazily allocates the registry's auth block so the username and password
// columns populate spec.auth, leaving it nil when neither is given.
func ensureAuth(reg *types.Registry) *types.Auth {
	if reg.Spec.Auth == nil {
		reg.Spec.Auth = &types.Auth{}
	}
	return reg.Spec.Auth
}

// trackedSkillStream renders a schema-valid installed Skill document for track.yaml.
// Pure (name in, bytes out) so it is unit-tested without the runtime.
func trackedSkillStream(name string) []byte {
	return []byte("apiVersion: " + types.APIVersion + "\n" +
		"kind: " + types.KindSkill + "\n" +
		"metadata:\n" +
		"  name: " + name + "\n" +
		"spec:\n" +
		"  digest: sha256:seed\n" +
		"  path: skills/" + name + "\n" +
		"  installedAt: \"2026-06-21T07:30:00Z\"\n" +
		"  updatedAt: \"2026-06-21T07:30:00Z\"\n")
}
