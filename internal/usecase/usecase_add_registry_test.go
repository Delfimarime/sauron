package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
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
	testName    = "acme"
	testFSURI   = "/srv"
	testHTTPURI = "https://acme.example"

	transportFilesystem = "filesystem"
	transportHTTP       = "http"
	transportGit        = "git"
)

// fixture bundles the use case and its mocked collaborators.
type fixture struct {
	uc         *AddRegistryUseCase
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

	f.uc = NewAddRegistryUseCase(AddRegistryUseCaseParams{
		Filesystem: f.filesystem,
		Git:        f.git,
		HTTP:       f.http,
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

func TestAddRegistryUseCase_Execute_Failures(t *testing.T) {
	t.Run("bad name yields usage", func(t *testing.T) {
		f := newFixture()
		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: "Bad_Name", URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeUsage)
		f.filesystem.AssertNotCalled(t, "Validate", mock.Anything)
	})

	t.Run("literal password yields usage", func(t *testing.T) {
		f := newFixture()
		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testHTTPURI, Transport: transportHTTP,
			Password: "literal-secret",
		})
		requireErrType(t, err, TypeUsage)
		f.http.AssertNotCalled(t, "Validate", mock.Anything)
	})

	t.Run("unknown transport yields usage", func(t *testing.T) {
		f := newFixture()
		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: "x", Transport: "ftp",
		})
		requireErrType(t, err, TypeUsage)
	})

	t.Run("ref on non-git transport yields usage", func(t *testing.T) {
		f := newFixture()
		f.http.On("Validate", mock.Anything).
			Return(fmt.Errorf("%w: ref unsupported", api.ErrUsage))

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testHTTPURI, Transport: transportHTTP, Ref: "main",
		})
		requireErrType(t, err, TypeUsage)
	})

	t.Run("existing name yields conflict", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).
			Return(&types.Registry{}, nil)

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeConflict)
	})

	t.Run("store lookup error yields io", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).
			Return(nil, errors.New("disk gone"))

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeIO)
	})

	t.Run("unset env var yields unreachable", func(t *testing.T) {
		// ACME_USER is deliberately left unset in the process environment.
		f := newFixture()
		f.http.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testHTTPURI, Transport: transportHTTP,
			Username: "${env:ACME_USER}",
		})
		ucErr := asUseCaseError(t, err, TypeUnreachable)
		assert.Contains(t, ucErr.Reason, "ACME_USER")
	})

	t.Run("open runtime error yields unreachable", func(t *testing.T) {
		f := newFixture()
		f.http.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)
		f.http.On("Open", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("%w: dial tcp", api.ErrRuntime))

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testHTTPURI, Transport: transportHTTP,
		})
		requireErrType(t, err, TypeUnreachable)
	})

	t.Run("open usage error yields usage", func(t *testing.T) {
		f := newFixture()
		f.http.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)
		f.http.On("Open", mock.Anything, mock.Anything).
			Return(nil, fmt.Errorf("%w: bad uri", api.ErrUsage))

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testHTTPURI, Transport: transportHTTP,
		})
		requireErrType(t, err, TypeUsage)
	})

	t.Run("validate non-usage error yields unreachable", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(errors.New("weird"))

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeUnreachable)
	})

	t.Run("list error yields unreachable", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)
		f.filesystem.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
		f.fs.On("List", mock.Anything, ".skills", mock.Anything).
			Return(nil, errors.New("io"))

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeUnreachable)
	})

	t.Run("empty presence scan yields unreachable hosts no artifact", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)
		f.filesystem.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
		f.stampAbsent()

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testFSURI, Transport: transportFilesystem,
		})
		ucErr := asUseCaseError(t, err, TypeUnreachable)
		assert.Equal(t, "hosts no artifact", ucErr.Reason)
	})

	t.Run("persist error yields io", func(t *testing.T) {
		f := newFixture()
		f.filesystem.On("Validate", mock.Anything).Return(nil)
		f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)
		f.filesystem.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
		f.stampPresent()
		f.store.On("Add", mock.Anything, mock.Anything).Return(errors.New("full"))

		err := f.uc.Execute(&AddRegistryRequest{
			Context: context.Background(), out: &bytes.Buffer{},
			Name: testName, URI: testFSURI, Transport: transportFilesystem,
		})
		requireErrType(t, err, TypeIO)
	})
}

func TestAddRegistryUseCase_Execute_HappyPath_Git(t *testing.T) {
	// Arrange.
	t.Setenv("GIT_USER", "real-user")
	t.Setenv("GIT_PASS", "real-pass")
	f := newFixture()
	f.git.On("Validate", mock.Anything).Return(nil)
	f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)

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
	f.store.On("Add", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			stored = args.Get(1).(types.Registry)
		}).
		Return(nil)

	out := &bytes.Buffer{}

	// Act.
	err := f.uc.Execute(&AddRegistryRequest{
		Context: context.Background(), out: out,
		Name: testName, URI: "git@host:repo.git", Transport: transportGit, Ref: "main",
		Username: "${env:GIT_USER}", Password: "${env:GIT_PASS}",
		Timeout: 15 * time.Second,
	})

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, types.TransportGit, stored.Spec.Transport)
	assert.Equal(t, "git@host:repo.git", stored.Spec.URI)
	assert.Equal(t, "main", stored.Spec.Ref)
	assert.Equal(t, "15s", stored.Spec.Timeout)
	require.NotNil(t, stored.Spec.Auth)
	assert.Equal(t, "${env:GIT_USER}", stored.Spec.Auth.Username)
	assert.Equal(t, "${env:GIT_PASS}", stored.Spec.Auth.Password)
	// Connecting uses the resolved values, never the references.
	assert.Equal(t, "real-user", openOpts.Username)
	assert.Equal(t, "real-pass", openOpts.Password)
	assert.Equal(t, "registered registry \"acme\" (git)\n", out.String())
}

func TestAddRegistryUseCase_Execute_HappyPath_NonGitDropsRef(t *testing.T) {
	// Arrange.
	f := newFixture()
	f.http.On("Validate", mock.Anything).Return(nil)
	f.store.On("FindByName", mock.Anything, testName).Return(nil, nil)
	f.http.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)
	f.stampPresent()

	var stored types.Registry
	f.store.On("Add", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			stored = args.Get(1).(types.Registry)
		}).
		Return(nil)

	out := &bytes.Buffer{}

	// Act: Ref is supplied but must be dropped for a non-git transport.
	err := f.uc.Execute(&AddRegistryRequest{
		Context: context.Background(), out: out,
		Name: testName, URI: testHTTPURI, Transport: transportHTTP, Ref: "ignored",
		SkipTLSVerify: true, CACert: "/ca.pem",
	})

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, stored.Spec.Ref)
	require.NotNil(t, stored.Spec.TLS)
	assert.True(t, stored.Spec.TLS.SkipVerify)
	assert.Equal(t, "/ca.pem", stored.Spec.TLS.CACert)
	assert.Nil(t, stored.Spec.Auth)
	assert.Equal(t, "registered registry \"acme\" (http)\n", out.String())
}
