package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// openFixture bundles the action and its mocked transport adapters.
type openFixture struct {
	action     *OpenRegistryUseCase
	filesystem *extension.MockBasedRegistry
	git        *extension.MockBasedRegistry
	http       *extension.MockBasedRegistry
	fs         *source.MockBasedFileSystem
}

// newOpenFixture wires an action over fresh mocks with an injected env lookup.
func newOpenFixture(env map[string]string) *openFixture {
	f := &openFixture{
		filesystem: &extension.MockBasedRegistry{},
		git:        &extension.MockBasedRegistry{},
		http:       &extension.MockBasedRegistry{},
		fs:         &source.MockBasedFileSystem{},
	}

	f.action = NewOpenRegistryUseCase(OpenRegistryUseCaseParams{
		Filesystem: f.filesystem,
		Git:        f.git,
		HTTP:       f.http,
		Logger:     zap.NewNop(),
	})
	f.action.lookupEnv = func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}

	return f
}

func TestOpenRegistryUseCase_Execute_TransportSelection(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// transport is the registry's configured transport.
		transport types.Transport
		// adapter selects the mock expected to be opened.
		adapter func(*openFixture) *extension.MockBasedRegistry
	}{
		{
			name:      "filesystem transport opens the filesystem adapter",
			transport: types.TransportFilesystem,
			adapter:   func(f *openFixture) *extension.MockBasedRegistry { return f.filesystem },
		},
		{
			name:      "git transport opens the git adapter",
			transport: types.TransportGit,
			adapter:   func(f *openFixture) *extension.MockBasedRegistry { return f.git },
		},
		{
			name:      "http transport opens the http adapter",
			transport: types.TransportHTTP,
			adapter:   func(f *openFixture) *extension.MockBasedRegistry { return f.http },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			f := newOpenFixture(nil)
			adapter := tt.adapter(f)
			adapter.On("Open", mock.Anything, mock.Anything).Return(f.fs, nil)

			// Act.
			got, err := f.action.Execute(context.Background(), types.Registry{
				Spec: types.RegistrySpec{Transport: tt.transport, Source: testFSURI},
			})

			// Assert: the selected adapter is opened and its file system returned.
			require.NoError(t, err)
			assert.Same(t, f.fs, got)
			adapter.AssertCalled(t, "Open", mock.Anything, mock.Anything)
		})
	}
}

func TestOpenRegistryUseCase_Execute_UnknownTransport(t *testing.T) {
	// Arrange.
	f := newOpenFixture(nil)

	// Act.
	_, err := f.action.Execute(context.Background(), types.Registry{
		Spec: types.RegistrySpec{Transport: types.Transport("ftp"), Source: "x"},
	})

	// Assert: a usage error names the unknown transport; no adapter is opened.
	ucErr := asUseCaseError(t, err, TypeUsage)
	assert.Contains(t, ucErr.Reason, "ftp")
	f.filesystem.AssertNotCalled(t, "Open", mock.Anything, mock.Anything)
}

func TestOpenRegistryUseCase_Execute_CredentialReferences(t *testing.T) {
	t.Run("set references resolve before open", func(t *testing.T) {
		// Arrange.
		f := newOpenFixture(map[string]string{"GIT_USER": "real-user", "GIT_PASS": "real-pass"})
		var opened extension.Options
		f.git.On("Open", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				for _, opt := range args.Get(1).([]extension.Option) {
					opt(&opened)
				}
			}).
			Return(f.fs, nil)

		// Act.
		_, err := f.action.Execute(context.Background(), types.Registry{
			Spec: types.RegistrySpec{
				Transport: types.TransportGit, Source: "git@host:repo.git", Revision: testRef,
				Credentials: &types.Credentials{Username: "${env:GIT_USER}", Password: "${env:GIT_PASS}"},
			},
		})

		// Assert: connecting uses the resolved values, never the references.
		require.NoError(t, err)
		assert.Equal(t, "real-user", opened.Username)
		assert.Equal(t, "real-pass", opened.Password)
		assert.Equal(t, testRef, opened.Ref)
	})

	t.Run("literal credentials, ssh key and tls pass through unchanged", func(t *testing.T) {
		// Arrange.
		f := newOpenFixture(nil)
		var opened extension.Options
		f.http.On("Open", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				for _, opt := range args.Get(1).([]extension.Option) {
					opt(&opened)
				}
			}).
			Return(f.fs, nil)

		// Act.
		_, err := f.action.Execute(context.Background(), types.Registry{
			Spec: types.RegistrySpec{
				Transport: types.TransportHTTP, Source: testHTTPURI, Timeout: "15s", SSHKey: "/key",
				Credentials: &types.Credentials{Username: "literal-user", Password: "literal-pass"},
				TLS:  &types.TLS{SkipVerify: true, CACert: "/ca.pem", ClientCert: "/c.pem", ClientKey: "/k.pem"},
			},
		})

		// Assert: literals are carried verbatim and every option is applied.
		require.NoError(t, err)
		assert.Equal(t, "literal-user", opened.Username)
		assert.Equal(t, "literal-pass", opened.Password)
		assert.Equal(t, "/key", opened.SSHKey)
		assert.True(t, opened.SkipTLSVerify)
		assert.Equal(t, "/ca.pem", opened.CACert)
		assert.Equal(t, "/c.pem", opened.ClientCert)
		assert.Equal(t, "/k.pem", opened.ClientKey)
	})

	t.Run("unset reference yields unreachable before open", func(t *testing.T) {
		// Arrange: ACME_USER is absent from the injected environment.
		f := newOpenFixture(nil)

		// Act.
		_, err := f.action.Execute(context.Background(), types.Registry{
			Spec: types.RegistrySpec{
				Transport: types.TransportHTTP, Source: testHTTPURI,
				Credentials: &types.Credentials{Username: "${env:ACME_USER}"},
			},
		})

		// Assert: the unset variable is unreachable and the source is never opened.
		ucErr := asUseCaseError(t, err, TypeUnreachable)
		assert.Contains(t, ucErr.Reason, "ACME_USER")
		f.http.AssertNotCalled(t, "Open", mock.Anything, mock.Anything)
	})
}

func TestOpenRegistryUseCase_Execute_OpenFailureClassification(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// openErr is the adapter's open failure.
		openErr error
		// want is the classified error type.
		want Type
	}{
		{
			name:    "runtime open failure is unreachable",
			openErr: fmt.Errorf("%w: dial tcp", api.ErrRuntime),
			want:    TypeUnreachable,
		},
		{
			name:    "opaque open failure is unreachable",
			openErr: errors.New("boom"),
			want:    TypeUnreachable,
		},
		{
			name:    "usage open failure stays usage",
			openErr: fmt.Errorf("%w: bad uri", api.ErrUsage),
			want:    TypeUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			f := newOpenFixture(nil)
			f.http.On("Open", mock.Anything, mock.Anything).Return(nil, tt.openErr)

			// Act.
			_, err := f.action.Execute(context.Background(), types.Registry{
				Spec: types.RegistrySpec{Transport: types.TransportHTTP, Source: testHTTPURI},
			})

			// Assert.
			requireErrType(t, err, tt.want)
		})
	}
}

func TestOpenRegistryUseCase_Execute_InvalidTimeout(t *testing.T) {
	// Arrange.
	f := newOpenFixture(nil)

	// Act.
	_, err := f.action.Execute(context.Background(), types.Registry{
		Spec: types.RegistrySpec{Transport: types.TransportHTTP, Source: testHTTPURI, Timeout: "nope"},
	})

	// Assert: an unparsable duration is a usage error and no source is opened.
	ucErr := asUseCaseError(t, err, TypeUsage)
	assert.Contains(t, ucErr.Reason, "nope")
	f.http.AssertNotCalled(t, "Open", mock.Anything, mock.Anything)
}
