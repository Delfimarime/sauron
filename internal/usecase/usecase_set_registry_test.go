package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const (
	testFSURI   = "/srv"
	testHTTPURI = "https://acme.example"

	transportFilesystem = "filesystem"
	transportHTTP       = "http"
	transportGit        = "git"

	testRef = "main"
)

// fixture bundles the use case and its mocked collaborators.
type fixture struct {
	uc         *SetRegistryUseCase
	filesystem *extension.MockBasedRegistry
	git        *extension.MockBasedRegistry
	http       *extension.MockBasedRegistry
	store      *storage.MockBasedRegistriesStore
	fs         *source.MockBasedFileSystem
}

// newFixture wires a use case over fresh mocks.
func newFixture() *fixture {
	f := &fixture{
		filesystem: &extension.MockBasedRegistry{},
		git:        &extension.MockBasedRegistry{},
		http:       &extension.MockBasedRegistry{},
		store:      &storage.MockBasedRegistriesStore{},
		fs:         &source.MockBasedFileSystem{},
	}

	open := NewOpenRegistryUseCase(OpenRegistryUseCaseParams{
		Filesystem: f.filesystem,
		Git:        f.git,
		HTTP:       f.http,
		Logger:     zap.NewNop(),
	})

	f.uc = NewSetRegistryUseCase(SetRegistryUseCaseParams{
		Filesystem: f.filesystem,
		Git:        f.git,
		HTTP:       f.http,
		Open:       open,
		Registries: f.store,
		Logger:     zap.NewNop(),
	})

	return f
}

// fileStub is a minimal source.File used to make a listing non-empty.
type fileStub struct {
	source.File
}

// stampPresent makes the filesystem report one artifact under the first root.
func (f *fixture) stampPresent() {
	f.fs.On("List", mock.Anything, ".skills", mock.Anything).
		Return([]source.File{fileStub{}}, nil)
}

// stampAbsent makes both artifact roots report no entries.
func (f *fixture) stampAbsent() {
	f.fs.On("List", mock.Anything, ".skills", mock.Anything).
		Return([]source.File{}, nil)
	f.fs.On("List", mock.Anything, ".agents", mock.Anything).
		Return([]source.File{}, nil)
}

// asUseCaseError asserts err is a *Error with the expected Type and returns it.
func asUseCaseError(t *testing.T, err error, want Type) *Error {
	t.Helper()
	require.Error(t, err)

	var ucErr *Error
	require.True(t, errors.As(err, &ucErr), "want *Error, got %T", err)
	assert.Equal(t, want, ucErr.Type)

	return ucErr
}

// requireErrType asserts err is a *Error with the expected Type.
func requireErrType(t *testing.T, err error, want Type) {
	t.Helper()
	_ = asUseCaseError(t, err, want)
}

func TestSetRegistryUseCase_Execute_Failures(t *testing.T) {
	t.Run("literal password yields usage", func(t *testing.T) {
		f := newFixture()
		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testHTTPURI, Transport: transportHTTP,
			Password: "literal-secret",
		})
		requireErrType(t, err, TypeUsage)
		f.http.AssertNotCalled(t, "Validate", mock.Anything)
	})

	t.Run("unknown transport yields usage", func(t *testing.T) {
		f := newFixture()
		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: "x", Transport: "ftp",
		})
		requireErrType(t, err, TypeUsage)
	})

	t.Run("ref on non-git transport yields usage", func(t *testing.T) {
		f := newFixture()
		f.http.On("Validate", mock.Anything).
			Return(fmt.Errorf("%w: ref unsupported", api.ErrUsage))

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testHTTPURI, Transport: transportHTTP, Ref: testRef,
		})
		requireErrType(t, err, TypeUsage)
	})

	t.Run("unset env var yields unreachable", func(t *testing.T) {
		// ACME_USER is deliberately left unset in the process environment.
		f := newFixture()
		f.http.On("Validate", mock.Anything).Return(nil)

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testHTTPURI, Transport: transportHTTP,
			Username: "${env:ACME_USER}",
		})
		ucErr := asUseCaseError(t, err, TypeUnreachable)
		assert.Contains(t, ucErr.Reason, "ACME_USER")
	})

	t.Run("open runtime error yields unreachable", func(t *testing.T) {
		f := newFixture()
		f.http.On("Validate", mock.Anything).Return(nil)
		f.http.On("Open", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("%w: dial tcp", api.ErrRuntime))

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testHTTPURI, Transport: transportHTTP,
		})
		requireErrType(t, err, TypeUnreachable)
	})

	t.Run("open usage error yields usage", func(t *testing.T) {
		f := newFixture()
		f.http.On("Validate", mock.Anything).Return(nil)
		f.http.On("Open", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("%w: bad uri", api.ErrUsage))

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testHTTPURI, Transport: transportHTTP,
		})
		requireErrType(t, err, TypeUsage)
	})

	t.Run("validate non-usage error yields unreachable", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(errors.New("weird"))

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeUnreachable)
	})

	t.Run("list error yields unreachable", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.filesystem.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
		f.fs.On("List", mock.Anything, ".skills", mock.Anything).
			Return(nil, errors.New("io"))

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeUnreachable)
	})

	t.Run("empty presence scan yields unreachable hosts no artifact", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.filesystem.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
		f.stampAbsent()

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testFSURI, Transport: transportFilesystem,
		})
		ucErr := asUseCaseError(t, err, TypeUnreachable)
		assert.Equal(t, "hosts no artifact", ucErr.Reason)
	})

	t.Run("persist error yields io", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.filesystem.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
		f.stampPresent()
		f.store.On("Set", mock.Anything, mock.Anything).Return(errors.New("full"))

		_, err := f.uc.Execute(context.Background(), SetRegistryInput{
			URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeIO)
	})
}

func TestSetRegistryUseCase_Execute_HappyPath_Git(t *testing.T) {
	// Arrange.
	t.Setenv("GIT_USER", "real-user")
	t.Setenv("GIT_PASS", "real-pass")
	f := newFixture()
	f.git.On("Validate", mock.Anything).Return(nil)

	var openOpts extension.Options
	f.git.On("Open", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			for _, opt := range args.Get(1).([]extension.Option) {
				opt(&openOpts)
			}
		}).
		Return(f.fs, nil)
	f.fs.On("List", mock.Anything, ".agents", mock.Anything).
		Return([]source.File{fileStub{}}, nil)
	f.fs.On("List", mock.Anything, ".skills", mock.Anything).
		Return([]source.File{}, nil)

	var stored types.Registry
	f.store.On("Set", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			stored = args.Get(1).(types.Registry)
		}).
		Return(nil)

	// Act. Run inside a synctest bubble so time.Now() is a fixed, controllable
	// instant; the use case stamps the audit timestamps from it.
	var err error
	var result *SetRegistryResult
	var want string
	synctest.Test(t, func(*testing.T) {
		want = time.Now().UTC().Format(time.RFC3339)
		result, err = f.uc.Execute(context.Background(), SetRegistryInput{
			URI: "git@host:repo.git", Transport: transportGit, Ref: testRef,
			Username: "${env:GIT_USER}", Password: "${env:GIT_PASS}",
			Timeout: 15 * time.Second,
		})
	})

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, types.TransportGit, stored.Spec.Transport)
	assert.Equal(t, "git@host:repo.git", stored.Spec.Source)
	assert.Equal(t, testRef, stored.Spec.Revision)
	assert.Equal(t, "15s", stored.Spec.Timeout)
	require.NotNil(t, stored.Spec.Credentials)
	assert.Equal(t, "${env:GIT_USER}", stored.Spec.Credentials.Username)
	assert.Equal(t, "${env:GIT_PASS}", stored.Spec.Credentials.Password)
	// Connecting uses the resolved values, never the references.
	assert.Equal(t, "real-user", openOpts.Username)
	assert.Equal(t, "real-pass", openOpts.Password)
	// Both audit timestamps are stamped with the current instant, equal on create.
	assert.Equal(t, want, stored.Metadata.CreatedAt)
	assert.Equal(t, want, stored.Metadata.LastUpdatedAt)
	// The result carries the configured URI and transport for the client to render.
	require.NotNil(t, result)
	assert.Equal(t, "git@host:repo.git", result.URI)
	assert.Equal(t, types.TransportGit, result.Transport)
}

func TestSetRegistryUseCase_Execute_HappyPath_NonGitDropsRef(t *testing.T) {
	// Arrange.
	f := newFixture()
	f.http.On("Validate", mock.Anything).Return(nil)
	f.http.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
	f.stampPresent()

	var stored types.Registry
	f.store.On("Set", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			stored = args.Get(1).(types.Registry)
		}).
		Return(nil)

	// Act: Ref is supplied but must be dropped for a non-git transport.
	result, err := f.uc.Execute(context.Background(), SetRegistryInput{
		URI: testHTTPURI, Transport: transportHTTP, Ref: "ignored",
		SkipTLSVerify: true, CACert: "/ca.pem",
	})

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, stored.Spec.Revision)
	require.NotNil(t, stored.Spec.TLS)
	assert.True(t, stored.Spec.TLS.SkipVerify)
	assert.Equal(t, "/ca.pem", stored.Spec.TLS.CACert)
	assert.Nil(t, stored.Spec.Credentials)
	require.NotNil(t, result)
	assert.Equal(t, testHTTPURI, result.URI)
	assert.Equal(t, types.TransportHTTP, result.Transport)
}
