//go:build unit

package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManifestsSplitsOnFileDirective(t *testing.T) {
	body := "# file: skills/go-style.yaml\n" +
		"apiVersion: sauron.raitonbl.com/v1\n" +
		"kind: Skill\n" +
		"metadata:\n" +
		"  name: go-style\n" +
		"# file: agents/code-reviewer.yml\n" +
		"apiVersion: sauron.raitonbl.com/v1\n" +
		"kind: Agent\n"

	resources, err := parseManifests(body)
	require.NoError(t, err)
	require.Len(t, resources, 2)

	assert.Equal(t, "skills/go-style.yaml", resources[0].Path)
	assert.Contains(t, string(resources[0].Content), "name: go-style")
	assert.NotContains(t, string(resources[0].Content), "Agent", "a file's content stops at the next directive")

	assert.Equal(t, "agents/code-reviewer.yml", resources[1].Path)
	assert.Contains(t, string(resources[1].Content), "kind: Agent")
}

func TestParseManifestsRejectsContentBeforeFirstDirectiveAndEmptyBody(t *testing.T) {
	_, err := parseManifests("kind: Skill\n# file: skills/go.yaml\n")
	assert.Error(t, err, "content before the first directive is rejected")

	_, err = parseManifests("   \n\n")
	assert.Error(t, err, "a doc-string with no directive is rejected")
}
