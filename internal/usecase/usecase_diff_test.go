package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// newDiffUseCase wires the diff use case over a fresh mocked track store,
// verifying its expectations on cleanup so an unused stub fails the test.
func newDiffUseCase(t *testing.T, track *storage.MockBasedTrackStore) *DiffUseCase {
	t.Helper()
	t.Cleanup(func() { track.AssertExpectations(t) })
	return NewDiffUseCase(DiffUseCaseParams{Track: track})
}

// trackedSkill builds a recorded skill at the given version for diff assertions.
func trackedSkill(name, version string) types.Artifact {
	return skillArtifact(name, version, "2024-01-01T00:00:00Z")
}

// TestDiffUseCase_Execute_Add categorizes a desired artifact that is not tracked
// as an addition, leaving the other groups empty.
func TestDiffUseCase_Execute_Add(t *testing.T) {
	// Arrange.
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return([]types.Artifact(nil), nil)
	uc := newDiffUseCase(t, track)
	desired := DesiredArtifact{Kind: types.KindSkill, Name: installSkillName, Version: "v1"}

	// Act.
	diff, err := uc.Execute(context.Background(), DiffRequest{Desired: []DesiredArtifact{desired}})

	// Assert.
	require.NoError(t, err)
	require.Len(t, diff.Add, 1)
	assert.Equal(t, desired, diff.Add[0])
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Unchanged)
	assert.Empty(t, diff.Remove)
}

// TestDiffUseCase_Execute_Update categorizes a tracked artifact whose desired
// version differs as an update.
func TestDiffUseCase_Execute_Update(t *testing.T) {
	// Arrange.
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return([]types.Artifact{trackedSkill(installSkillName, "old")}, nil)
	uc := newDiffUseCase(t, track)
	desired := DesiredArtifact{Kind: types.KindSkill, Name: installSkillName, Version: "new"}

	// Act.
	diff, err := uc.Execute(context.Background(), DiffRequest{Desired: []DesiredArtifact{desired}})

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, diff.Add)
	require.Len(t, diff.Update, 1)
	assert.Equal(t, desired, diff.Update[0].Desired)
	assert.Equal(t, trackedSkill(installSkillName, "old"), diff.Update[0].Prior)
	assert.Empty(t, diff.Unchanged)
	assert.Empty(t, diff.Remove)
}

// TestDiffUseCase_Execute_Unchanged categorizes a tracked artifact whose desired
// version is equal as unchanged, carrying the recorded artifact.
func TestDiffUseCase_Execute_Unchanged(t *testing.T) {
	// Arrange.
	prior := trackedSkill(installSkillName, "v1")
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return([]types.Artifact{prior}, nil)
	uc := newDiffUseCase(t, track)
	desired := DesiredArtifact{Kind: types.KindSkill, Name: installSkillName, Version: "v1"}

	// Act.
	diff, err := uc.Execute(context.Background(), DiffRequest{Desired: []DesiredArtifact{desired}})

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, diff.Add)
	assert.Empty(t, diff.Update)
	require.Len(t, diff.Unchanged, 1)
	assert.Equal(t, prior, diff.Unchanged[0])
	assert.Empty(t, diff.Remove)
}

// TestDiffUseCase_Execute_RemovalWhenIncluded marks a tracked artifact absent
// from the desired set for removal only when removals are included.
func TestDiffUseCase_Execute_RemovalWhenIncluded(t *testing.T) {
	// Arrange: orphan is tracked but not desired.
	orphan := trackedSkill("orphan", "v1")
	kept := trackedSkill(installSkillName, "v1")
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return([]types.Artifact{kept, orphan}, nil)
	uc := newDiffUseCase(t, track)
	desired := DesiredArtifact{Kind: types.KindSkill, Name: installSkillName, Version: "v1"}

	// Act.
	diff, err := uc.Execute(context.Background(), DiffRequest{
		Desired: []DesiredArtifact{desired}, IncludeRemovals: true,
	})

	// Assert: orphan removed, go-style unchanged, no additions or updates.
	require.NoError(t, err)
	assert.Empty(t, diff.Add)
	assert.Empty(t, diff.Update)
	require.Len(t, diff.Unchanged, 1)
	assert.Equal(t, kept, diff.Unchanged[0])
	require.Len(t, diff.Remove, 1)
	assert.Equal(t, orphan, diff.Remove[0])
}

// TestDiffUseCase_Execute_PartialDesiredKeepsSiblings proves a partial desired set
// does NOT mark untouched tracked siblings for removal when removals are excluded.
func TestDiffUseCase_Execute_PartialDesiredKeepsSiblings(t *testing.T) {
	// Arrange: sibling is tracked but not in the partial desired set.
	sibling := trackedSkill("sibling", "v1")
	kept := trackedSkill(installSkillName, "v1")
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return([]types.Artifact{kept, sibling}, nil)
	uc := newDiffUseCase(t, track)
	desired := DesiredArtifact{Kind: types.KindSkill, Name: installSkillName, Version: "v1"}

	// Act: IncludeRemovals defaults to false.
	diff, err := uc.Execute(context.Background(), DiffRequest{Desired: []DesiredArtifact{desired}})

	// Assert: no removals despite the untouched sibling; no additions or updates.
	require.NoError(t, err)
	assert.Empty(t, diff.Add)
	assert.Empty(t, diff.Update)
	require.Len(t, diff.Unchanged, 1)
	assert.Equal(t, kept, diff.Unchanged[0])
	assert.Empty(t, diff.Remove)
}

// TestDiffUseCase_Execute_KeyedByKindAndName distinguishes a skill and an agent
// sharing a name, so each desired entry diffs against its own kind.
func TestDiffUseCase_Execute_KeyedByKindAndName(t *testing.T) {
	// Arrange: a skill named shared is tracked; an agent named shared is not.
	skill := trackedSkill("shared", "v1")
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return([]types.Artifact{skill}, nil)
	uc := newDiffUseCase(t, track)

	// Act.
	diff, err := uc.Execute(context.Background(), DiffRequest{Desired: []DesiredArtifact{
		{Kind: types.KindSkill, Name: "shared", Version: "v1"},
		{Kind: types.KindAgent, Name: "shared", Version: "v1"},
	}})

	// Assert: the skill is unchanged, the agent is an addition; no updates or removals.
	require.NoError(t, err)
	require.Len(t, diff.Unchanged, 1)
	assert.Equal(t, types.KindSkill, diff.Unchanged[0].Kind)
	require.Len(t, diff.Add, 1)
	assert.Equal(t, types.KindAgent, diff.Add[0].Kind)
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Remove)
}

// TestDiffUseCase_Execute_Empty returns an all-empty diff for empty inputs.
func TestDiffUseCase_Execute_Empty(t *testing.T) {
	// Arrange.
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return([]types.Artifact(nil), nil)
	uc := newDiffUseCase(t, track)

	// Act.
	diff, err := uc.Execute(context.Background(), DiffRequest{IncludeRemovals: true})

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, diff.Add)
	assert.Empty(t, diff.Update)
	assert.Empty(t, diff.Unchanged)
	assert.Empty(t, diff.Remove)
}

// TestDiffUseCase_Execute_TrackReadError surfaces a track read failure as io.
func TestDiffUseCase_Execute_TrackReadError(t *testing.T) {
	// Arrange.
	track := &storage.MockBasedTrackStore{}
	track.On("List", mock.Anything).Return(nil, errors.New("boom"))
	uc := newDiffUseCase(t, track)

	// Act.
	_, err := uc.Execute(context.Background(), DiffRequest{})

	// Assert.
	requireErrType(t, err, TypeIO)
}
