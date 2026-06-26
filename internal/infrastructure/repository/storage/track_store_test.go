package storage

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// validSkillYAML is a schema-valid Skill document.
const validSkillYAML = `apiVersion: sauron.raitonbl.com/v1
kind: Skill
metadata:
  name: sauron-acme-go-style
spec:
  digest: sha256:aaa
  path: sauron-acme-go-style
  installedAt: "2024-01-01T00:00:00Z"
  updatedAt: "2024-01-01T00:00:00Z"
`

// validAgentYAML is a schema-valid Agent document.
const validAgentYAML = `apiVersion: sauron.raitonbl.com/v1
kind: Agent
metadata:
  name: sauron-acme-code-reviewer
spec:
  digest: sha256:bbb
  path: sauron-acme-code-reviewer
  installedAt: "2024-01-01T00:00:00Z"
  updatedAt: "2024-01-01T00:00:00Z"
`

// tsZero is the seeded artifact timestamp shared by the track-store tests.
const tsZero = "2024-01-01T00:00:00Z"

// newTestTrackStore builds a TrackStore over an in-memory filesystem.
func newTestTrackStore(t *testing.T) (TrackStore, *Store) {
	t.Helper()
	store, _ := newTestStore(t)
	return NewTrackStore(store), store
}

// TestTrackStoreListAbsent returns an empty slice when track.yaml is absent.
func TestTrackStoreListAbsent(t *testing.T) {
	// Arrange.
	track, _ := newTestTrackStore(t)

	// Act.
	artifacts, err := track.List(context.Background())

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, artifacts)
}

// TestTrackStoreListDecodesMixedStream decodes the Skill and Agent documents from
// a single mixed-kind track.yaml stream into a discriminated Artifact slice.
func TestTrackStoreListDecodesMixedStream(t *testing.T) {
	// Arrange: a stream holding one Skill and one Agent.
	track, store := newTestTrackStore(t)
	stream := []byte(documentSeparator + validSkillYAML + documentSeparator + validAgentYAML)
	require.NoError(t, afero.WriteFile(store.fs, trackFile, stream, filePerm))

	// Act.
	artifacts, err := track.List(context.Background())

	// Assert: skills come first, then agents, each carrying its document kind.
	require.NoError(t, err)
	require.Len(t, artifacts, 2)
	assert.Equal(t, "sauron-acme-go-style", artifacts[0].Metadata.Name)
	assert.Equal(t, types.KindSkill, artifacts[0].Kind)
	assert.Equal(t, "sauron-acme-go-style", artifacts[0].Spec.Path)
	assert.Equal(t, "sauron-acme-code-reviewer", artifacts[1].Metadata.Name)
	assert.Equal(t, types.KindAgent, artifacts[1].Kind)
}

// TestTrackStoreListPropagatesReadError surfaces a read failure.
func TestTrackStoreListPropagatesReadError(t *testing.T) {
	// Arrange: a malformed stream makes the underlying read fail.
	track, store := newTestTrackStore(t)
	require.NoError(t, afero.WriteFile(store.fs, trackFile, []byte("\tnot: [valid"), filePerm))

	// Act.
	_, err := track.List(context.Background())

	// Assert.
	require.Error(t, err)
}

// TestTrackStoreUpdateRoundTrips persists an updated artifact and reads it back,
// asserting the bumped updatedAt survives and the document keeps its kind.
func TestTrackStoreUpdateRoundTrips(t *testing.T) {
	// Arrange: one Skill already recorded.
	track, store := newTestTrackStore(t)
	require.NoError(t, afero.WriteFile(store.fs, trackFile, []byte(documentSeparator+validSkillYAML), filePerm))

	// Act: bump the skill's updatedAt through Update.
	updated := types.Artifact{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindSkill},
		Metadata: types.Metadata{Name: "sauron-acme-go-style"},
		Spec: types.ArtifactSpec{
			Digest:      "sha256:aaa",
			Path:        "sauron-acme-go-style",
			InstalledAt: tsZero,
			UpdatedAt:   "2024-02-02T00:00:00Z",
		},
	}
	require.NoError(t, track.Update(context.Background(), updated))

	// Assert: exactly one skill remains, updatedAt bumped, kind preserved.
	artifacts, err := track.List(context.Background())
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	assert.Equal(t, types.KindSkill, artifacts[0].Kind)
	assert.Equal(t, "2024-02-02T00:00:00Z", artifacts[0].Spec.UpdatedAt)
}

// TestTrackStoreUpdateAppendsWhenAbsent adds a new artifact when none matches.
func TestTrackStoreUpdateAppendsWhenAbsent(t *testing.T) {
	// Arrange.
	track, _ := newTestTrackStore(t)
	added := types.Artifact{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindAgent},
		Metadata: types.Metadata{Name: "sauron-new-agent"},
		Spec: types.ArtifactSpec{
			Digest:      "sha256:ccc",
			Path:        "sauron-new-agent",
			InstalledAt: tsZero,
			UpdatedAt:   tsZero,
		},
	}

	// Act.
	require.NoError(t, track.Update(context.Background(), added))

	// Assert.
	artifacts, err := track.List(context.Background())
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	assert.Equal(t, "sauron-new-agent", artifacts[0].Metadata.Name)
	assert.Equal(t, types.KindAgent, artifacts[0].Kind)
}

// TestTrackStoreUpdateRejectsUnknownKind errors when the artifact carries no
// recognized document kind.
func TestTrackStoreUpdateRejectsUnknownKind(t *testing.T) {
	// Arrange.
	track, _ := newTestTrackStore(t)

	// Act.
	err := track.Update(context.Background(), types.Artifact{
		TypeMeta: types.TypeMeta{Kind: "Bogus"},
		Metadata: types.Metadata{Name: "x"},
	})

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}

// TestMockBasedTrackStore exercises the testify mock.
func TestMockBasedTrackStore(t *testing.T) {
	// Arrange.
	m := &MockBasedTrackStore{}
	ctx := context.Background()
	want := []types.Artifact{{TypeMeta: types.TypeMeta{Kind: types.KindSkill}, Metadata: types.Metadata{Name: "s"}}}
	m.On("List", ctx).Return(want, nil)
	m.On("Update", ctx, types.Artifact{}).Return(nil)

	// Act.
	artifacts, listErr := m.List(ctx)
	updateErr := m.Update(ctx, types.Artifact{})

	// Assert.
	require.NoError(t, listErr)
	require.NoError(t, updateErr)
	assert.Equal(t, want, artifacts)
	m.AssertExpectations(t)
}

// TestMockBasedTrackStoreNil returns a nil slice without panicking.
func TestMockBasedTrackStoreNil(t *testing.T) {
	// Arrange.
	m := &MockBasedTrackStore{}
	ctx := context.Background()
	m.On("List", ctx).Return(nil, nil)

	// Act.
	artifacts, err := m.List(ctx)

	// Assert.
	require.NoError(t, err)
	assert.Nil(t, artifacts)
}

// TestNewTrackStoreType confirms the constructor returns the interface.
func TestNewTrackStoreType(t *testing.T) {
	// Arrange.
	fs := afero.NewMemMapFs()
	store, err := NewStore(fs)
	require.NoError(t, err)

	// Act.
	track := NewTrackStore(store)

	// Assert.
	require.NotNil(t, track)
}
