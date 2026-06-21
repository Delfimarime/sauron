package gherkin

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// seedColumns are the table columns the seed step understands; name carries the
// metadata.name, the username/password columns populate spec.auth, and every
// other column populates the matching spec field.
var seedColumns = map[string]struct{}{
	"name": {}, "transport": {}, "uri": {}, "ref": {}, "timeout": {},
	"username": {}, "password": {}, "sshKey": {},
	"creationTimestamp": {}, "lastUpdatedTimestamp": {},
}

// listController owns the list-registries Given seeds (arranging the read-only
// listing per the graybox-arrange exception). Like every controller it holds only
// the runtime; the seed logic lives in pure helpers so it is unit-tested without a
// process or real fs.
type listController struct {
	rt runtime.Runtime
}

func (c *listController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the following registries are configured:$`, c.registriesConfigured)
	sc.Step(`^the registries file contains:$`, c.registriesFileContains)
}

// registriesConfigured seeds a schema-valid Registry document stream into
// registries.yaml under $SAURON_HOME.
func (c *listController) registriesConfigured(ctx context.Context, table *godog.Table) error {
	stream, err := buildRegistryStream(table)
	if err != nil {
		return err
	}
	return c.rt.CopyTo(ctx, registriesFile, stream)
}

// registriesFileContains writes the exact bytes the scenario provides into
// registries.yaml, so the file under test — including a deliberately malformed one
// (FR-006) — is designed explicitly by the feature, never synthesized in code.
func (c *listController) registriesFileContains(ctx context.Context, content *godog.DocString) error {
	return c.rt.CopyTo(ctx, registriesFile, []byte(content.Content))
}

// buildRegistryStream turns a |name|transport|uri|…| table into a multi-document
// YAML stream of schema-valid Registry documents. Pure (table in, bytes out) so it
// is unit-tested without the runtime.
func buildRegistryStream(table *godog.Table) ([]byte, error) {
	if table == nil || len(table.Rows) < 2 {
		return nil, fmt.Errorf("the following registries are configured: needs a header and at least one row")
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
			return nil, fmt.Errorf("encode registry %q: %w", reg.Metadata.Name, err)
		}
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// registryFromRow maps one table row to a schema-valid Registry, stamping the
// apiVersion/kind envelope and populating the spec fields the columns carry.
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
	if reg.Metadata.Name == "" {
		return types.Registry{}, fmt.Errorf("a name column value is required")
	}
	return reg, nil
}

// ensureAuth lazily allocates the registry's auth block so the username and
// password columns populate spec.auth, leaving it nil when neither is given.
func ensureAuth(reg *types.Registry) *types.Auth {
	if reg.Spec.Auth == nil {
		reg.Spec.Auth = &types.Auth{}
	}
	return reg.Spec.Auth
}

// nameColumn reads the name column down the data rows of a rendered list: the first
// whitespace-delimited token of every line after the header. Pure (text in, names
// out) so it is unit-tested without a process.
func nameColumn(output string) []string {
	var names []string
	headerSeen := false
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if !headerSeen {
			headerSeen = true // the first non-empty line is the NAME … header row
			continue
		}
		names = append(names, strings.Fields(line)[0])
	}
	return names
}

// expectedOrder splits the step argument into the expected name sequence; names are
// separated by commas and/or whitespace.
func expectedOrder(list string) []string {
	return strings.FieldsFunc(list, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
}
