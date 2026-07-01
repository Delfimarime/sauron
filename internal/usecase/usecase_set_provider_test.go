package usecase

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// Shared artifact-name fixtures across the set-provider and migrate tests.
const (
	artifactSkillName = "sauron-acme-go-style"
	artifactAgentName = "sauron-acme-code-reviewer"
)

// providerFixture bundles the use case and its mocked collaborators.
type providerFixture struct {
	uc        *SetProviderUseCase
	providers *storage.MockBasedProvidersStore
	migrate   *MockBasedUseCase[MigrateRequest, MigrateResponse]
}

// newProviderFixture wires a use case over fresh mocks, verifying their
// expectations on cleanup so an unused or over-specified stub fails the test.
func newProviderFixture(t *testing.T) *providerFixture {
	t.Helper()
	f := &providerFixture{
		providers: &storage.MockBasedProvidersStore{},
		migrate:   &MockBasedUseCase[MigrateRequest, MigrateResponse]{},
	}
	f.uc = NewSetProviderUseCase(SetProviderUseCaseParams{
		Providers: f.providers,
		Migrate:   f.migrate,
		Logger:    zap.NewNop(),
	})
	t.Cleanup(func() {
		f.providers.AssertExpectations(t)
		f.migrate.AssertExpectations(t)
	})
	return f
}

// TestSetProviderUseCase_Execute_Failures covers the classified error paths.
func TestSetProviderUseCase_Execute_Failures(t *testing.T) {
	t.Run("unknown name yields usage", func(t *testing.T) {
		f := newProviderFixture(t)
		_, err := f.uc.Execute(context.Background(), SetProviderRequest{Provider: "bogus"})
		requireErrType(t, err, TypeUsage)
		f.providers.AssertNotCalled(t, "Get", mock.Anything)
		f.providers.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
	})

	t.Run("read error yields io", func(t *testing.T) {
		f := newProviderFixture(t)
		f.providers.On("Get", mock.Anything).Return(nil, errors.New("io"))
		_, err := f.uc.Execute(context.Background(), SetProviderRequest{Provider: types.ProviderClaude})
		requireErrType(t, err, TypeIO)
	})

	t.Run("migrate error propagates", func(t *testing.T) {
		f := newProviderFixture(t)
		f.providers.On("Get", mock.Anything).Return(
			&types.Provider{Metadata: types.Metadata{Name: types.ProviderClaude}}, nil,
		)
		f.migrate.On("Execute", mock.Anything, mock.Anything).Return(nil, NewIOError("disk"))
		_, err := f.uc.Execute(context.Background(), SetProviderRequest{Provider: types.ProviderZencoder})
		requireErrType(t, err, TypeIO)
		f.providers.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
	})

	t.Run("persist error yields io", func(t *testing.T) {
		f := newProviderFixture(t)
		f.providers.On("Get", mock.Anything).Return(nil, nil)
		f.providers.On("Set", mock.Anything, mock.Anything).Return(errors.New("full"))
		_, err := f.uc.Execute(context.Background(), SetProviderRequest{Provider: types.ProviderClaude})
		requireErrType(t, err, TypeIO)
		f.migrate.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	})
}

// TestSetProviderUseCase_Execute_FirstSet persists a new provider with no
// migration (there is no current provider to switch from) and equal audit stamps.
func TestSetProviderUseCase_Execute_FirstSet(t *testing.T) {
	// Arrange.
	f := newProviderFixture(t)
	f.providers.On("Get", mock.Anything).Return(nil, nil)

	var stored types.Provider
	f.providers.On("Set", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			stored = args.Get(1).(types.Provider)
		}).
		Return(nil)

	// Act. Run inside a synctest bubble so time.Now() is fixed and controllable.
	var err error
	var result *SetProviderResponse
	var want string
	synctest.Test(t, func(*testing.T) {
		want = time.Now().UTC().Format(time.RFC3339)
		result, err = f.uc.Execute(context.Background(), SetProviderRequest{Provider: types.ProviderZencoder})
	})

	// Assert: no migration runs on a first set; the plan is empty.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.ProviderZencoder, result.Provider)
	assert.False(t, result.Unchanged)
	assert.Equal(t, 0, result.Migrated)
	assert.Empty(t, result.Skills)
	assert.Empty(t, result.Agents)
	assert.Equal(t, types.ProviderZencoder, stored.Metadata.Name)
	assert.Equal(t, want, stored.Metadata.CreatedAt)
	assert.Equal(t, want, stored.Metadata.LastUpdatedAt)
	f.migrate.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

// TestSetProviderUseCase_Execute_Switch migrates the installed set and groups the
// moved artifacts by type, then persists the new provider.
func TestSetProviderUseCase_Execute_Switch(t *testing.T) {
	// Arrange: claude is active; a switch to zencoder migrates two artifacts.
	f := newProviderFixture(t)
	f.providers.On("Get", mock.Anything).Return(
		&types.Provider{Metadata: types.Metadata{Name: types.ProviderClaude}}, nil,
	)

	var in MigrateRequest
	f.migrate.On("Execute", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			in = args.Get(1).(MigrateRequest)
		}).
		Return(&MigrateResponse{Moved: []types.Artifact{
			{TypeMeta: types.TypeMeta{Kind: types.KindSkill}, Metadata: types.Metadata{Name: artifactSkillName}},
			{TypeMeta: types.TypeMeta{Kind: types.KindAgent}, Metadata: types.Metadata{Name: artifactAgentName}},
		}}, nil)
	f.providers.On("Set", mock.Anything, mock.Anything).Return(nil)

	// Act.
	result, err := f.uc.Execute(context.Background(), SetProviderRequest{Provider: types.ProviderZencoder})

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.ProviderClaude, in.From)
	assert.Equal(t, types.ProviderZencoder, in.To)
	assert.Equal(t, 2, result.Migrated)
	assert.Equal(t, []string{artifactSkillName}, result.Skills)
	assert.Equal(t, []string{artifactAgentName}, result.Agents)
}

// TestSetProviderUseCase_Execute_SwitchPersistsDespitePartialFailure records the
// new provider even when some migration steps failed (FR-005).
func TestSetProviderUseCase_Execute_SwitchPersistsDespitePartialFailure(t *testing.T) {
	// Arrange: one artifact moved, one failed.
	f := newProviderFixture(t)
	f.providers.On("Get", mock.Anything).Return(
		&types.Provider{Metadata: types.Metadata{Name: types.ProviderClaude}}, nil,
	)
	f.migrate.On("Execute", mock.Anything, mock.Anything).Return(&MigrateResponse{
		Moved: []types.Artifact{
			{TypeMeta: types.TypeMeta{Kind: types.KindSkill}, Metadata: types.Metadata{Name: artifactSkillName}},
		},
		Failures: []MigrateFailure{
			{Artifact: types.Artifact{TypeMeta: types.TypeMeta{Kind: types.KindAgent}, Metadata: types.Metadata{Name: "sauron-x"}}, Reason: "missing"},
		},
	}, nil)

	var stored types.Provider
	f.providers.On("Set", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			stored = args.Get(1).(types.Provider)
		}).
		Return(nil)

	// Act.
	result, err := f.uc.Execute(context.Background(), SetProviderRequest{Provider: types.ProviderZencoder})

	// Assert: provider recorded; only the moved artifact appears in the plan.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.ProviderZencoder, stored.Metadata.Name)
	assert.Equal(t, 1, result.Migrated)
	assert.Equal(t, []string{artifactSkillName}, result.Skills)
	assert.Empty(t, result.Agents)

	// The stranded artifact is surfaced, not silently dropped.
	require.Len(t, result.Failures, 1)
	assert.Equal(t, "sauron-x", result.Failures[0].Artifact.Metadata.Name)
	assert.Equal(t, "missing", result.Failures[0].Reason)
}

// TestSetProviderUseCase_Execute_Unchanged returns unchanged and persists nothing
// when the requested provider is already active.
func TestSetProviderUseCase_Execute_Unchanged(t *testing.T) {
	// Arrange: claude already configured.
	f := newProviderFixture(t)
	f.providers.On("Get", mock.Anything).Return(
		&types.Provider{Metadata: types.Metadata{Name: types.ProviderClaude}}, nil,
	)

	// Act.
	result, err := f.uc.Execute(context.Background(), SetProviderRequest{Provider: types.ProviderClaude})

	// Assert: no migration, no write.
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, types.ProviderClaude, result.Provider)
	assert.True(t, result.Unchanged)
	assert.Equal(t, 0, result.Migrated)
	f.providers.AssertNotCalled(t, "Set", mock.Anything, mock.Anything)
	f.migrate.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}
