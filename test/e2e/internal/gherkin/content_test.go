//go:build unit

package gherkin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTree materializes a path->content map under a fresh temp dir and returns its
// root. Writes land in the test temp dir only, never the real filesystem.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o600))
	}
	return root
}

func TestCollectResources(t *testing.T) {
	root := writeTree(t, map[string]string{
		".skills/go-style/skill.yaml":      "kind: Skill",
		".agents/code-reviewer/agent.yaml": "kind: Agent",
		".DS_Store":                        "noise",
	})

	resources, err := collectResources(root)
	require.NoError(t, err)

	got := map[string]string{}
	for _, r := range resources {
		got[r.Path] = string(r.Content)
	}
	assert.Equal(t, "kind: Skill", got[".skills/go-style/skill.yaml"], "nested file served at its slash path")
	assert.Equal(t, "kind: Agent", got[".agents/code-reviewer/agent.yaml"])
	assert.NotContains(t, got, ".DS_Store", "macOS noise is skipped")
}

func TestCollectResourcesErrors(t *testing.T) {
	_, err := collectResources(filepath.Join(t.TempDir(), "absent"))
	assert.Error(t, err, "a missing directory is an error")

	file := filepath.Join(t.TempDir(), "a-file")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o600))
	_, err = collectResources(file)
	assert.Error(t, err, "a file is not a directory; use the file step")
}

func TestFileResource(t *testing.T) {
	root := writeTree(t, map[string]string{"manifest.yaml": "kind: Skill"})
	path := filepath.Join(root, "manifest.yaml")

	r, err := fileResource(path, ".skills/go/skill.yaml")
	require.NoError(t, err)
	assert.Equal(t, ".skills/go/skill.yaml", r.Path)
	assert.Equal(t, "kind: Skill", string(r.Content))

	// An empty served path defaults to the file's base name.
	r, err = fileResource(path, "")
	require.NoError(t, err)
	assert.Equal(t, "manifest.yaml", r.Path)

	_, err = fileResource(filepath.Join(root, "absent.yaml"), "x")
	assert.Error(t, err)
}

func TestExposeDirectoryFeedsTheSource(t *testing.T) {
	root := writeTree(t, map[string]string{".skills/go/skill.yaml": "kind: Skill"})
	src := &fakeSource{}

	require.NoError(t, exposeDirectory(src, root))
	require.Len(t, src.exposed, 1)
	assert.Equal(t, ".skills/go/skill.yaml", src.exposed[0].Path)
}

func TestExposeFileFeedsTheSource(t *testing.T) {
	root := writeTree(t, map[string]string{"skill.yaml": "kind: Skill"})
	src := &fakeSource{}

	require.NoError(t, exposeFile(src, filepath.Join(root, "skill.yaml"), ".skills/go/skill.yaml"))
	require.Len(t, src.exposed, 1)
	assert.Equal(t, ".skills/go/skill.yaml", src.exposed[0].Path)

	// exposeFile surfaces loader errors rather than panicking.
	assert.Error(t, exposeFile(&fakeSource{}, filepath.Join(root, "absent"), "x"))
}
