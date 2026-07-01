package storage

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// otherSkillYAML is a second schema-valid Skill document.
const otherSkillYAML = `apiVersion: sauron.raitonbl.com/v1
kind: Skill
metadata:
  name: sauron-other
spec:
  version: sha256:zzz
  path: sauron-other
  installedAt: "2024-01-01T00:00:00Z"
  updatedAt: "2024-01-01T00:00:00Z"
`

// replacementSkillYAML replaces sauron-acme-go-style with a bumped updatedAt.
const replacementSkillYAML = `apiVersion: sauron.raitonbl.com/v1
kind: Skill
metadata:
  name: sauron-acme-go-style
spec:
  version: sha256:new
  path: sauron-acme-go-style
  installedAt: "2024-01-01T00:00:00Z"
  updatedAt: "2024-03-03T00:00:00Z"
`

// TestStoreUpsertReplacesMatching replaces the document sharing a (kind, name)
// while preserving every other document in the file.
func TestStoreUpsertReplacesMatching(t *testing.T) {
	// Arrange: two skills recorded.
	store, _ := newTestStore(t)
	require.NoError(t, store.Append(context.Background(), types.KindSkill, nodeFromYAML(t, validSkillYAML)))
	require.NoError(t, store.Append(context.Background(), types.KindSkill, nodeFromYAML(t, otherSkillYAML)))

	// Act: upsert a replacement for the first skill.
	require.NoError(t, store.Upsert(
		context.Background(), types.KindSkill, "sauron-acme-go-style",
		nodeFromYAML(t, replacementSkillYAML),
	))

	// Assert: both skills remain; the replacement took effect.
	docs, err := store.FindAll(context.Background(), types.KindSkill)
	require.NoError(t, err)
	require.Len(t, docs, 2)
}

// TestStoreUpsertUnknownKind reports an error for a kind with no backing file.
func TestStoreUpsertUnknownKind(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	err := store.Upsert(context.Background(), "Nonexistent", "x", nodeFromYAML(t, validSkillYAML))

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}

// TestStoreUpsertPropagatesReadError surfaces a malformed-stream read failure.
func TestStoreUpsertPropagatesReadError(t *testing.T) {
	// Arrange: a malformed track.yaml.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, trackFile, []byte("\tnot: [valid"), filePerm))

	// Act.
	err := store.Upsert(context.Background(), types.KindSkill, "x", nodeFromYAML(t, validSkillYAML))

	// Assert.
	require.Error(t, err)
	assert.NotErrorIs(t, err, errUnknownKind)
}
