//go:build unit

package gherkin

import (
	"testing"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// table builds a godog.Table from a header and data rows, matching the shape godog
// hands the seed step (so the pure builders are exercised exactly as in a scenario).
func table(rows ...[]string) *godog.Table {
	out := &godog.Table{}
	for _, row := range rows {
		cells := make([]*messages.PickleTableCell, len(row))
		for i, value := range row {
			cells[i] = &messages.PickleTableCell{Value: value}
		}
		out.Rows = append(out.Rows, &messages.PickleTableRow{Cells: cells})
	}
	return out
}

func TestBuildRegistryStreamStampsEnvelopeAndSpec(t *testing.T) {
	stream, err := buildRegistryStream(table(
		[]string{"name", "transport", "uri"},
		[]string{"acme", "git", "git@github.com:acme/artifacts.git"},
		[]string{"internal", "http", "https://reg.example.com/"},
	))
	require.NoError(t, err)

	// The seeded bytes round-trip through the same decoder the assertions use,
	// proving the stream is the schema-valid form a user could author.
	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 2)

	assert.Equal(t, types.APIVersion, regs[0].APIVersion)
	assert.Equal(t, types.KindRegistry, regs[0].Kind)
	assert.Equal(t, "acme", regs[0].Metadata.Name)
	assert.Equal(t, types.TransportGit, regs[0].Spec.Transport)
	assert.Equal(t, "git@github.com:acme/artifacts.git", regs[0].Spec.URI)
	assert.Equal(t, "internal", regs[1].Metadata.Name)
	assert.Equal(t, types.TransportHTTP, regs[1].Spec.Transport)
}

func TestBuildRegistryStreamCarriesOptionalColumns(t *testing.T) {
	stream, err := buildRegistryStream(table(
		[]string{"name", "transport", "uri", "ref", "timeout", "sshKey", "creationTimestamp", "lastUpdatedTimestamp"},
		[]string{"acme", "git", "git@github.com:acme/artifacts.git", "v1.2.0", "45s", "/home/dev/.ssh/id_ed25519", "2026-06-21T07:30:00Z", "2026-06-22T08:00:00Z"},
	))
	require.NoError(t, err)

	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 1)
	assert.Equal(t, "v1.2.0", regs[0].Spec.Ref)
	assert.Equal(t, "45s", regs[0].Spec.Timeout)
	assert.Equal(t, "/home/dev/.ssh/id_ed25519", regs[0].Spec.SSHKey)
	assert.Equal(t, "2026-06-21T07:30:00Z", regs[0].Metadata.CreationTimestamp)
	assert.Equal(t, "2026-06-22T08:00:00Z", regs[0].Metadata.LastUpdatedTimestamp)
}

func TestBuildRegistryStreamRejectsUnknownColumnAndEmptyTable(t *testing.T) {
	_, err := buildRegistryStream(table(
		[]string{"name", "bogus"},
		[]string{"acme", "x"},
	))
	assert.Error(t, err)

	_, err = buildRegistryStream(table([]string{"name", "transport", "uri"}))
	assert.Error(t, err, "a header with no data rows is rejected")
}

func TestNameColumnReadsFirstTokenAfterHeader(t *testing.T) {
	rendered := "NAME      TRANSPORT  URI\n" +
		"acme      git        git@github.com:acme/artifacts.git\n" +
		"internal  http       https://reg.example.com/\n"
	assert.Equal(t, []string{"acme", "internal"}, nameColumn(rendered))

	assert.Nil(t, nameColumn(""), "no output yields no names")
	assert.Nil(t, nameColumn("NAME  TRANSPORT  URI\n"), "header only yields no names")
}

func TestExpectedOrderSplitsOnCommaAndSpace(t *testing.T) {
	assert.Equal(t, []string{"internal", "acme"}, expectedOrder("internal, acme"))
	assert.Equal(t, []string{"acme"}, expectedOrder("acme"))
	assert.Equal(t, []string{"a", "b", "c"}, expectedOrder("a b,c"))
}
