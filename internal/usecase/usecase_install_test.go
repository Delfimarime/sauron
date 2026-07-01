package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// shared literals across the install assertions.
const (
	installSkillName = "go-style"
	installSkillFile = "SKILL.md"
	installAgentName = "code-reviewer"
	treeVersion      = "tree-v1"
)

// fakeFile is a concrete source.File backed by an in-memory body, used so the
// install assertions exercise the real Read/write path without mocking the
// reader plumbing on every entry. As a listing entry only its Name and Version
// are read; as a fetched entry its body is read.
type fakeFile struct {
	name    string
	version string
	body    string
	dir     bool
}

func (f fakeFile) Name() string      { return f.name }
func (f fakeFile) IsDirectory() bool { return f.dir }
func (f fakeFile) Size() int64       { return int64(len(f.body)) }
func (f fakeFile) Version() string   { return f.version }
func (f fakeFile) Read(_ context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.body)), nil
}

// installFixture bundles the install use case over an in-memory provider
// filesystem and mocked collaborators, including a mocked diff step.
type installFixture struct {
	uc         *InstallUseCase
	providers  *storage.MockBasedProvidersStore
	registries *storage.MockBasedRegistriesStore
	track      *storage.MockBasedTrackStore
	open       *MockBasedOpenRegistryUseCase
	diff       *MockBasedUseCase[DiffRequest, DiffResponse]
	fs         *source.MockBasedFileSystem
	providerFs afero.Fs
}

// newInstallFixture wires the use case over fresh mocks and a memory-backed
// provider filesystem, registering mock-expectation verification on cleanup so an
// unused or over-specified stub fails the test.
func newInstallFixture(t *testing.T) *installFixture {
	t.Helper()
	providers := &storage.MockBasedProvidersStore{}
	registries := &storage.MockBasedRegistriesStore{}
	track := &storage.MockBasedTrackStore{}
	open := &MockBasedOpenRegistryUseCase{}
	diff := &MockBasedUseCase[DiffRequest, DiffResponse]{}
	fs := &source.MockBasedFileSystem{}
	providerFs := afero.NewMemMapFs()

	t.Cleanup(func() {
		providers.AssertExpectations(t)
		registries.AssertExpectations(t)
		track.AssertExpectations(t)
		open.AssertExpectations(t)
		diff.AssertExpectations(t)
		fs.AssertExpectations(t)
	})

	return &installFixture{
		uc: NewInstallUseCase(InstallUseCaseParams{
			Logger:     zap.NewNop(),
			Providers:  providers,
			Registries: registries,
			Track:      track,
			Open:       open,
			Diff:       diff,
			Fs:         providerFs,
		}),
		providers:  providers,
		registries: registries,
		track:      track,
		open:       open,
		diff:       diff,
		fs:         fs,
		providerFs: providerFs,
	}
}

// withProvider stubs the active provider as claude.
func (f *installFixture) withProvider() {
	f.providers.On("Get", mock.Anything).
		Return(&types.Provider{Metadata: types.Metadata{Name: types.ProviderClaude}}, nil)
}

// withRegistry stubs registry resolution and open to the fixture's source.
func (f *installFixture) withRegistry() {
	f.registries.On("Get", mock.Anything).
		Return(&types.Registry{Spec: types.RegistrySpec{Source: "https://acme.example"}}, nil)
	f.open.On("Execute", mock.Anything, mock.Anything).Return(f.fs, nil)
}

// withListing stubs List for the given root to return the listing entries, from
// which the use case resolves desired versions cheaply.
func (f *installFixture) withListing(root string, entries ...source.File) {
	f.fs.On("List", mock.Anything, root, mock.Anything).Return(entries, nil)
}

// withListError stubs List for the given root to fail.
func (f *installFixture) withListError(root string, err error) {
	f.fs.On("List", mock.Anything, root, mock.Anything).Return(nil, err)
}

// withDiff stubs the diff step to return the given plan, capturing the input it
// was composed with so resolve can be asserted.
func (f *installFixture) withDiff(plan *DiffResponse, captured *DiffRequest) {
	call := f.diff.On("Execute", mock.Anything, mock.Anything)
	if captured != nil {
		call = call.Run(func(args mock.Arguments) { *captured = args.Get(1).(DiffRequest) })
	}
	call.Return(plan, nil)
}

// withDiffError stubs the diff step to fail.
func (f *installFixture) withDiffError(err error) {
	f.diff.On("Execute", mock.Anything, mock.Anything).Return(nil, err)
}

// withFetch stubs Fetch for the given root/name URI to return files.
func (f *installFixture) withFetch(uri string, files ...source.File) {
	f.fs.On("Fetch", mock.Anything, uri).Return(files, nil)
}

// captureUpdate stubs Update, capturing every persisted artifact.
func (f *installFixture) captureUpdate(captured *[]types.Artifact) {
	f.track.On("Update", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*captured = append(*captured, args.Get(1).(types.Artifact))
		}).
		Return(nil)
}

func (f *installFixture) run(in InstallRequest) (*InstallResponse, error) {
	return f.uc.Execute(context.Background(), in)
}

// skillArtifact builds a recorded skill at the given version.
func skillArtifact(name, version, installedAt string) types.Artifact {
	return types.Artifact{
		TypeMeta: types.TypeMeta{APIVersion: types.APIVersion, Kind: types.KindSkill},
		Metadata: types.Metadata{Name: name},
		Spec: types.ArtifactSpec{
			Version:     version,
			Path:        installPath(types.KindSkill, name),
			InstalledAt: installedAt,
			UpdatedAt:   installedAt,
		},
	}
}

// desiredSkill builds the desired entry the listing resolves for the test skill.
func desiredSkill(version string) DesiredArtifact {
	return DesiredArtifact{Kind: types.KindSkill, Name: installSkillName, Version: version}
}

// TestInstallUseCase_Execute_FreshAdd installs an unrecorded skill: it resolves
// the source version from the listing, diffs it to an addition, then fetches,
// writes, and records it under the active provider directory.
func TestInstallUseCase_Execute_FreshAdd(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListing("skills", fakeFile{name: installSkillName, version: treeVersion})
	var diffIn DiffRequest
	f.withDiff(&DiffResponse{Add: []DesiredArtifact{desiredSkill(treeVersion)}}, &diffIn)
	f.withFetch("skills/go-style",
		fakeFile{name: installSkillFile, version: treeVersion, body: "body"},
		fakeFile{name: "lib/util.go", version: treeVersion, body: "pkg lib"},
	)
	var updated []types.Artifact
	f.captureUpdate(&updated)

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert: the listing version reached the diff as a desired entry.
	require.NoError(t, err)
	assert.False(t, diffIn.IncludeRemovals)
	require.Len(t, diffIn.Desired, 1)
	assert.Equal(t, desiredSkill(treeVersion), diffIn.Desired[0])

	// Counted as added, with the recorded version and path.
	require.Len(t, result.Added, 1)
	assert.Empty(t, result.Updated)
	assert.Empty(t, result.Failures)

	added := result.Added[0]
	assert.Equal(t, types.KindSkill, added.Kind)
	assert.Equal(t, installSkillName, added.Metadata.Name)
	assert.Equal(t, treeVersion, added.Spec.Version)
	assert.Equal(t, "skills/sauron-go-style", added.Spec.Path)
	assert.NotEmpty(t, added.Spec.InstalledAt)
	assert.Equal(t, added.Spec.InstalledAt, added.Spec.UpdatedAt)

	// Persisted exactly once, with the same shape.
	require.Len(t, updated, 1)
	assert.Equal(t, added, updated[0])

	// Files written under the provider directory, artifact-relative paths kept.
	body, err := afero.ReadFile(f.providerFs, ".claude/skills/sauron-go-style/SKILL.md")
	require.NoError(t, err)
	assert.Equal(t, "body", string(body))
	nested, err := afero.ReadFile(f.providerFs, ".claude/skills/sauron-go-style/lib/util.go")
	require.NoError(t, err)
	assert.Equal(t, "pkg lib", string(nested))
}

// TestInstallUseCase_Execute_AgentPath installs an agent under the agents root
// and the agents provider subdir.
func TestInstallUseCase_Execute_AgentPath(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListing("agents", fakeFile{name: installAgentName, version: "v9"})
	f.withDiff(&DiffResponse{Add: []DesiredArtifact{
		{Kind: types.KindAgent, Name: installAgentName, Version: "v9"},
	}}, nil)
	f.withFetch("agents/"+installAgentName,
		fakeFile{name: "AGENT.md", version: "v9", body: "agent"},
	)
	var updated []types.Artifact
	f.captureUpdate(&updated)

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindAgent, Names: []string{installAgentName}})

	// Assert.
	require.NoError(t, err)
	require.Len(t, result.Added, 1)
	assert.Equal(t, "agents/sauron-code-reviewer", result.Added[0].Spec.Path)
	exists, _ := afero.Exists(f.providerFs, ".claude/agents/sauron-code-reviewer/AGENT.md")
	assert.True(t, exists)
}

// TestInstallUseCase_Execute_ChangedUpdate reconciles a recorded artifact the diff
// reports as an update: it keeps installedAt and bumps updatedAt.
func TestInstallUseCase_Execute_ChangedUpdate(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListing("skills", fakeFile{name: installSkillName, version: "new-v1"})
	prior := skillArtifact(installSkillName, "old-v0", "2024-01-01T00:00:00Z")
	f.withDiff(&DiffResponse{Update: []UpdatePlan{{Prior: prior, Desired: desiredSkill("new-v1")}}}, nil)
	f.withFetch("skills/go-style",
		fakeFile{name: installSkillFile, version: "new-v1", body: "fresh"},
	)
	var updated []types.Artifact
	f.captureUpdate(&updated)

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert: counted as updated, version bumped, installedAt preserved.
	require.NoError(t, err)
	assert.Empty(t, result.Added)
	require.Len(t, result.Updated, 1)

	got := result.Updated[0]
	assert.Equal(t, "new-v1", got.Spec.Version)
	assert.Equal(t, "2024-01-01T00:00:00Z", got.Spec.InstalledAt)
	assert.NotEqual(t, "2024-01-01T00:00:00Z", got.Spec.UpdatedAt)
	require.Len(t, updated, 1)
}

// TestInstallUseCase_Execute_UnchangedNoOp leaves an artifact the diff reports as
// unchanged untouched: no fetch, no rewrite, no track update, no count.
func TestInstallUseCase_Execute_UnchangedNoOp(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	prior := skillArtifact(installSkillName, "same-v1", "2024-01-01T00:00:00Z")
	f.withListing("skills", fakeFile{name: installSkillName, version: "same-v1"})
	f.withDiff(&DiffResponse{Unchanged: []types.Artifact{prior}}, nil)

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert: nothing changed, nothing fetched or written, track left alone.
	require.NoError(t, err)
	assert.Empty(t, result.Added)
	assert.Empty(t, result.Updated)
	assert.Empty(t, result.Failures)
	f.fs.AssertNotCalled(t, "Fetch", mock.Anything, mock.Anything)
	f.track.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	exists, _ := afero.Exists(f.providerFs, ".claude/skills/sauron-go-style/SKILL.md")
	assert.False(t, exists)
}

// TestInstallUseCase_Execute_NoProvider fails with a non-usage runtime error and
// writes nothing, resolving the provider before touching the registry (FR-005).
func TestInstallUseCase_Execute_NoProvider(t *testing.T) {
	// Arrange: no provider; registry/open are not stubbed and must not be hit.
	f := newInstallFixture(t)
	f.providers.On("Get", mock.Anything).Return(nil, nil)

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert.
	require.Nil(t, result)
	var useErr *Error
	require.ErrorAs(t, err, &useErr)
	assert.NotEqual(t, TypeUsage, useErr.Type, "maps to exit 1, not 2")
	f.registries.AssertNotCalled(t, "Get", mock.Anything)
	f.open.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

// TestInstallUseCase_Execute_UnofferedSibling records a name the registry does not
// offer as a failure while the offered sibling still installs (FR-006). The
// unoffered name is dropped from the desired set before the diff.
func TestInstallUseCase_Execute_UnofferedSibling(t *testing.T) {
	// Arrange: the listing offers go-style only.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListing("skills", fakeFile{name: installSkillName, version: "v1"})
	var diffIn DiffRequest
	f.withDiff(&DiffResponse{Add: []DesiredArtifact{desiredSkill("v1")}}, &diffIn)
	f.withFetch("skills/go-style",
		fakeFile{name: installSkillFile, version: "v1", body: "ok"},
	)
	var updated []types.Artifact
	f.captureUpdate(&updated)

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName, "ghost"}})

	// Assert: one installed, one reported, run continued; ghost never reached the diff.
	require.NoError(t, err)
	require.Len(t, result.Added, 1)
	assert.Equal(t, installSkillName, result.Added[0].Metadata.Name)
	require.Len(t, result.Failures, 1)
	assert.Equal(t, "ghost", result.Failures[0].Name)
	assert.NotEmpty(t, result.Failures[0].Reason)
	assert.False(t, result.Failures[0].Fatal, "a not-offered name is a benign skip (exit 0)")
	require.Len(t, diffIn.Desired, 1)
	assert.Equal(t, installSkillName, diffIn.Desired[0].Name)
}

// TestInstallUseCase_Execute_EmptyVersionSkipped reports and skips an artifact the
// listing declares with no version (versioning FR-005), before the diff.
func TestInstallUseCase_Execute_EmptyVersionSkipped(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListing("skills", fakeFile{name: installSkillName, version: ""})
	var diffIn DiffRequest
	f.withDiff(&DiffResponse{}, &diffIn)

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, result.Added)
	assert.Empty(t, result.Updated)
	require.Len(t, result.Failures, 1)
	assert.Equal(t, installSkillName, result.Failures[0].Name)
	assert.False(t, result.Failures[0].Fatal, "no declared version is a benign skip (exit 0)")
	assert.Empty(t, diffIn.Desired)
	f.track.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

// TestInstallUseCase_Execute_RegistryUnreachable propagates the open failure as a
// runtime error (FR-007).
func TestInstallUseCase_Execute_RegistryUnreachable(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.registries.On("Get", mock.Anything).
		Return(&types.Registry{Spec: types.RegistrySpec{Source: "https://acme.example"}}, nil)
	f.open.On("Execute", mock.Anything, mock.Anything).
		Return(nil, NewUnreachableError("source down"))

	// Act.
	_, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert.
	requireErrType(t, err, TypeUnreachable)
}

// TestInstallUseCase_Execute_ListUnreachable maps a listing failure to a runtime
// unreachable error.
func TestInstallUseCase_Execute_ListUnreachable(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListError("skills", errors.New("source down"))

	// Act.
	_, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert.
	requireErrType(t, err, TypeUnreachable)
}

// TestInstallUseCase_Execute_UnknownKind rejects a kind outside skills/agents with
// a usage error before any store is read.
func TestInstallUseCase_Execute_UnknownKind(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)

	// Act.
	_, err := f.run(InstallRequest{Kind: "Widget", Names: []string{installSkillName}})

	// Assert.
	requireErrType(t, err, TypeUsage)
	f.providers.AssertNotCalled(t, "Get", mock.Anything)
}

// TestInstallUseCase_Execute_ProviderReadError surfaces a provider read failure as
// io and reaches neither the registry nor the source.
func TestInstallUseCase_Execute_ProviderReadError(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.providers.On("Get", mock.Anything).Return(nil, errors.New("disk gone"))

	// Act.
	_, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert.
	requireErrType(t, err, TypeIO)
	f.registries.AssertNotCalled(t, "Get", mock.Anything)
}

// TestInstallUseCase_Execute_RegistryNotConfigured fails with not-found when no
// registry is configured.
func TestInstallUseCase_Execute_RegistryNotConfigured(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.registries.On("Get", mock.Anything).Return(nil, nil)

	// Act.
	_, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert.
	requireErrType(t, err, TypeNotFound)
}

// TestInstallUseCase_Execute_PersistFailureRecorded records a name whose track
// update fails as a per-name failure, leaving the run to continue (FR-006).
func TestInstallUseCase_Execute_PersistFailureRecorded(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListing("skills", fakeFile{name: installSkillName, version: "v1"})
	f.withDiff(&DiffResponse{Add: []DesiredArtifact{desiredSkill("v1")}}, nil)
	f.withFetch("skills/go-style",
		fakeFile{name: "docs", dir: true},
		fakeFile{name: installSkillFile, version: "v1", body: "ok"},
	)
	f.track.On("Update", mock.Anything, mock.Anything).Return(errors.New("write denied"))

	// Act.
	result, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert: nothing added, the name reported.
	require.NoError(t, err)
	assert.Empty(t, result.Added)
	require.Len(t, result.Failures, 1)
	assert.Equal(t, installSkillName, result.Failures[0].Name)
	assert.True(t, result.Failures[0].Fatal, "a track failure could not be persisted (exit 1)")
}

// TestInstallUseCase_Execute_DiffError surfaces a diff failure as a runtime error,
// installing nothing.
func TestInstallUseCase_Execute_DiffError(t *testing.T) {
	// Arrange.
	f := newInstallFixture(t)
	f.withProvider()
	f.withRegistry()
	f.withListing("skills", fakeFile{name: installSkillName, version: "v1"})
	f.withDiffError(NewIOError("read installed set"))

	// Act.
	_, err := f.run(InstallRequest{Kind: types.KindSkill, Names: []string{installSkillName}})

	// Assert.
	requireErrType(t, err, TypeIO)
	f.track.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
