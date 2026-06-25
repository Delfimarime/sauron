//go:build unit

package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

func TestParseManifestsSplitsOnFileDirective(t *testing.T) {
	body := "# file: .skills/go-style.yaml\n" +
		"apiVersion: sauron.raitonbl.com/v1\n" +
		"kind: Skill\n" +
		"metadata:\n" +
		"  name: go-style\n" +
		"# file: .personas/backend-dev.yml\n" +
		"apiVersion: sauron.raitonbl.com/v1\n" +
		"kind: Persona\n"

	resources, err := parseManifests(body)
	require.NoError(t, err)
	require.Len(t, resources, 2)

	assert.Equal(t, ".skills/go-style.yaml", resources[0].Path)
	assert.Contains(t, string(resources[0].Content), "name: go-style")
	assert.NotContains(t, string(resources[0].Content), "Persona", "a file's content stops at the next directive")

	assert.Equal(t, ".personas/backend-dev.yml", resources[1].Path)
	assert.Contains(t, string(resources[1].Content), "kind: Persona")
}

func TestParseManifestsRejectsContentBeforeFirstDirectiveAndEmptyBody(t *testing.T) {
	_, err := parseManifests("kind: Skill\n# file: .skills/go.yaml\n")
	assert.Error(t, err, "content before the first directive is rejected")

	_, err = parseManifests("   \n\n")
	assert.Error(t, err, "a doc-string with no directive is rejected")
}

func TestFilesystemRegistryStreamIsSchemaValid(t *testing.T) {
	stream, err := filesystemRegistryStream("acme", "/tmp/registry/acme")
	require.NoError(t, err)

	regs, err := decodeRegistries(stream)
	require.NoError(t, err)
	require.Len(t, regs, 1)

	assert.Equal(t, types.APIVersion, regs[0].APIVersion)
	assert.Equal(t, types.KindRegistry, regs[0].Kind)
	assert.Equal(t, "acme", regs[0].Metadata.Name)
	assert.Equal(t, types.TransportFilesystem, regs[0].Spec.Transport)
	assert.Equal(t, "/tmp/registry/acme", regs[0].Spec.URI)
}

func TestCatalogueHasRowMatchesOnFields(t *testing.T) {
	output := "NAME        KIND\n" +
		"go-style    skill\n" +
		"sql-review  skill\n" +
		"showing 1–2 (page 1, limit 20)\n"

	assert.True(t, catalogueHasRow(output, "go-style skill"))
	assert.True(t, catalogueHasRow(output, "sql-review skill"))
	assert.False(t, catalogueHasRow(output, "go-style agent"))
	assert.False(t, catalogueHasRow(output, "missing skill"))
}

func TestHasLineMatchesTrimmedLine(t *testing.T) {
	output := "NAME  KIND\ncode-reviewer  agent\nshowing 1–1 (page 1, limit 20)\n"

	assert.True(t, hasLine(output, "showing 1–1 (page 1, limit 20)"))
	assert.False(t, hasLine(output, "showing 0 results (page 1, limit 20)"))
}
