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
		[]string{"transport", "uri"},
		[]string{"git", "git@github.com:acme/artifacts.git"},
	))
	require.NoError(t, err)

	// The seeded bytes round-trip through the same decoder the assertions use,
	// proving the stream is the schema-valid form a user could author.
	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 1)

	assert.Equal(t, types.APIVersion, regs[0].APIVersion)
	assert.Equal(t, types.KindRegistry, regs[0].Kind)
	assert.Equal(t, types.TransportGit, regs[0].Spec.Transport)
	assert.Equal(t, "git@github.com:acme/artifacts.git", regs[0].Spec.URI)
}

func TestBuildRegistryStreamCarriesOptionalColumns(t *testing.T) {
	stream, err := buildRegistryStream(table(
		[]string{"transport", "uri", "ref", "timeout", "sshKey", "creationTimestamp", "lastUpdatedTimestamp"},
		[]string{"git", "git@github.com:acme/artifacts.git", "v1.2.0", "45s", "/home/dev/.ssh/id_ed25519", "2026-06-21T07:30:00Z", "2026-06-22T08:00:00Z"},
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
		[]string{"transport", "bogus"},
		[]string{"git", "x"},
	))
	assert.Error(t, err)

	_, err = buildRegistryStream(table([]string{"transport", "uri"}))
	assert.Error(t, err, "a header with no data rows is rejected")
}

func TestTrackedSkillStreamIsSchemaValid(t *testing.T) {
	skills, err := decodeSkills(trackedSkillStream("sauron-acme-go-style"))
	require.NoError(t, err)
	require.Len(t, skills, 1)
	assert.Equal(t, "sauron-acme-go-style", skills[0].Metadata.Name)
	assert.NotEmpty(t, skills[0].Spec.Digest)
	assert.NotEmpty(t, skills[0].Spec.Path)
}
