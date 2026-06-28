package usecase

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const (
	claudeDir   = ".claude"
	zencoderDir  = ".zencoder"
	skillPath   = "sauron-acme-go-style/skill.yaml"
	missingPath = "sauron-missing/skill.yaml"
)

// skillTrackDoc renders a schema-valid Skill document for track.yaml.
func skillTrackDoc(name, path string) string {
	return `---
apiVersion: sauron.raitonbl.com/v1
kind: Skill
metadata:
  name: ` + name + `
spec:
  version: sha256:abc
  path: ` + path + `
  installedAt: "2024-01-01T00:00:00Z"
  updatedAt: "2024-01-01T00:00:00Z"
`
}

// migrateFixture bundles the use case over a real track store and an isolated
// provider filesystem.
type migrateFixture struct {
	uc         *MigrateUseCase
	providerFs afero.Fs
	track      storage.TrackStore
}

// newMigrateFixture wires a MigrateUseCase over in-memory filesystems, seeding
// track.yaml with the supplied document stream.
func newMigrateFixture(t *testing.T, trackStream string) *migrateFixture {
	t.Helper()
	storageFs := afero.NewMemMapFs()
	if trackStream != "" {
		require.NoError(t, afero.WriteFile(storageFs, "track.yaml", []byte(trackStream), 0o644))
	}
	store, err := storage.NewStore(storageFs)
	require.NoError(t, err)
	track := storage.NewTrackStore(store)

	providerFs := afero.NewMemMapFs()
	uc := NewMigrateUseCase(MigrateUseCaseParams{Track: track, Fs: providerFs, Logger: zap.NewNop()})

	return &migrateFixture{uc: uc, providerFs: providerFs, track: track}
}

// TestMigrateUseCase_Execute_MovesArtifact relocates the installed artifact from
// the source provider directory to the destination and bumps updatedAt.
func TestMigrateUseCase_Execute_MovesArtifact(t *testing.T) {
	// Arrange: one skill recorded, its file present under .claude.
	f := newMigrateFixture(t, skillTrackDoc(artifactSkillName, skillPath))
	require.NoError(t, afero.WriteFile(f.providerFs, claudeDir+"/"+skillPath, []byte("body"), 0o644))

	// Act.
	result, err := f.uc.Execute(context.Background(), MigrateInput{From: types.ProviderClaude, To: types.ProviderZencoder})

	// Assert: moved, source gone, destination present.
	require.NoError(t, err)
	require.Len(t, result.Moved, 1)
	assert.Empty(t, result.Failures)

	srcExists, _ := afero.Exists(f.providerFs, claudeDir+"/"+skillPath)
	dstExists, _ := afero.Exists(f.providerFs, zencoderDir+"/"+skillPath)
	assert.False(t, srcExists, "source removed")
	assert.True(t, dstExists, "destination created")

	// updatedAt bumped away from the seeded value, recorded in the track file.
	assert.NotEqual(t, "2024-01-01T00:00:00Z", result.Moved[0].Spec.UpdatedAt)
	artifacts, err := f.track.List(context.Background())
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	assert.NotEqual(t, "2024-01-01T00:00:00Z", artifacts[0].Spec.UpdatedAt)
}

// TestMigrateUseCase_Execute_PartialFailure records the artifact whose source is
// missing as failed while still migrating the others (FR-005).
func TestMigrateUseCase_Execute_PartialFailure(t *testing.T) {
	// Arrange: two skills recorded; only the first has a source file.
	stream := skillTrackDoc(artifactSkillName, skillPath) +
		skillTrackDoc("sauron-missing", missingPath)
	f := newMigrateFixture(t, stream)
	require.NoError(t, afero.WriteFile(f.providerFs, claudeDir+"/"+skillPath, []byte("body"), 0o644))

	// Act.
	result, err := f.uc.Execute(context.Background(), MigrateInput{From: types.ProviderClaude, To: types.ProviderZencoder})

	// Assert: one moved, one failure recorded.
	require.NoError(t, err)
	require.Len(t, result.Moved, 1)
	require.Len(t, result.Failures, 1)
	assert.Equal(t, artifactSkillName, result.Moved[0].Metadata.Name)
	assert.Equal(t, "sauron-missing", result.Failures[0].Artifact.Metadata.Name)
}

// TestMigrateUseCase_Execute_EmptyTrack moves nothing and reports no failure when
// nothing is installed.
func TestMigrateUseCase_Execute_EmptyTrack(t *testing.T) {
	// Arrange.
	f := newMigrateFixture(t, "")

	// Act.
	result, err := f.uc.Execute(context.Background(), MigrateInput{From: types.ProviderClaude, To: types.ProviderZencoder})

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, result.Moved)
	assert.Empty(t, result.Failures)
}

// TestMigrateUseCase_Execute_ReadError surfaces a track read failure as io.
func TestMigrateUseCase_Execute_ReadError(t *testing.T) {
	// Arrange: a malformed track.yaml.
	f := newMigrateFixture(t, "\tnot: [valid")

	// Act.
	_, err := f.uc.Execute(context.Background(), MigrateInput{From: types.ProviderClaude, To: types.ProviderZencoder})

	// Assert.
	requireErrType(t, err, TypeIO)
}
