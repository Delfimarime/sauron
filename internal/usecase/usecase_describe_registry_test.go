package usecase

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// describe-test literals, named to satisfy goconst across the package.
const (
	describeName = "acme"
	userRef      = "${env:ACME_USER}"
	tokenRef     = "${env:ACME_TOKEN}"
	gitURI       = "git@github.com:acme/artifacts.git"

	labelName      = "name:"
	labelTransport = "transport:"
	labelURI       = "uri:"
	labelRef       = "ref:"
	labelAuth      = "auth:"
	labelTimeout   = "timeout:"
	labelTLS       = "tls:"
	labelSSHKey    = "sshKey:"
	labelCreated   = "creationTimestamp:"
	labelUpdated   = "lastUpdatedTimestamp:"

	createdStamp = "2026-06-21T07:30:00Z"
	updatedStamp = "2026-06-22T08:00:00Z"
)

// describeFixture bundles the describe use case and its mocked store.
type describeFixture struct {
	uc    *DescribeRegistryUseCase
	store *storage.MockBasedRegistriesStore
}

// newDescribeFixture wires a describe use case over a fresh store mock.
func newDescribeFixture() *describeFixture {
	store := &storage.MockBasedRegistriesStore{}
	return &describeFixture{
		store: store,
		uc: NewDescribeRegistryUseCase(DescribeRegistryUseCaseParams{
			Registries: store,
			Logger:     zap.NewNop(),
		}),
	}
}

// fullRegistry is a registry populated across every describable field.
func fullRegistry() *types.Registry {
	return &types.Registry{
		Metadata: types.Metadata{
			Name:                 describeName,
			CreationTimestamp:    createdStamp,
			LastUpdatedTimestamp: updatedStamp,
		},
		Spec: types.RegistrySpec{
			Transport: types.TransportGit,
			URI:       gitURI,
			Ref:       "v1.2.0",
			Auth: &types.Auth{
				Username: userRef,
				Password: tokenRef,
			},
			Timeout: "45s",
		},
	}
}

// TestDescribeRegistrySuccess covers the find → project → render pipeline across
// the default view, field projection, and the nested auth block (FR-001/002/003).
func TestDescribeRegistrySuccess(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// fields is the requested --fields selection.
		fields []string
		// registry is the stored record.
		registry *types.Registry
		// wantContains are substrings the descriptor must contain, in order.
		wantContains []string
		// wantAbsent are substrings the descriptor must never contain.
		wantAbsent []string
	}{
		{
			name:     "default shows every populated field",
			registry: fullRegistry(),
			wantContains: []string{
				labelName, describeName,
				labelTransport, string(types.TransportGit),
				labelURI, gitURI,
				labelRef, "v1.2.0",
				labelAuth,
				"username:", userRef,
				"password:", tokenRef,
				labelTimeout, "45s",
				labelCreated, createdStamp,
				labelUpdated, updatedStamp,
			},
		},
		{
			name:         "default omits unpopulated fields",
			registry:     &types.Registry{Metadata: types.Metadata{Name: describeName}, Spec: types.RegistrySpec{Transport: types.TransportGit, URI: "u"}},
			wantContains: []string{labelName, labelTransport, labelURI},
			wantAbsent:   []string{labelRef, labelAuth, labelTLS, labelSSHKey, labelTimeout, labelCreated, labelUpdated},
		},
		{
			name:         "fields projects and orders, name forced first",
			fields:       []string{fieldTransport, fieldURI},
			registry:     fullRegistry(),
			wantContains: []string{labelName, labelTransport, labelURI},
			wantAbsent:   []string{labelRef, labelAuth, labelTimeout},
		},
		{
			name:         "auth renders the stored env references, never a secret",
			fields:       []string{fieldAuth},
			registry:     fullRegistry(),
			wantContains: []string{labelAuth, userRef, tokenRef},
			wantAbsent:   []string{"s3cr3t"},
		},
		{
			name:   "tls and sshKey render their populated sub-fields",
			fields: []string{fieldTLS, fieldSSHKey},
			registry: &types.Registry{
				Metadata: types.Metadata{Name: describeName},
				Spec: types.RegistrySpec{
					Transport: types.TransportHTTP,
					SSHKey:    "/home/dev/.ssh/id_ed25519",
					TLS: &types.TLS{
						SkipVerify: true,
						CACert:     "/etc/ssl/ca.pem",
						ClientCert: "/etc/ssl/client.pem",
						ClientKey:  "/etc/ssl/client.key",
					},
				},
			},
			wantContains: []string{
				labelTLS,
				"skipVerify: true",
				"caCert:", "/etc/ssl/ca.pem",
				"clientCert:", "/etc/ssl/client.pem",
				"clientKey:", "/etc/ssl/client.key",
				labelSSHKey, "/home/dev/.ssh/id_ed25519",
			},
		},
		{
			name:   "an empty tls block is omitted",
			fields: []string{fieldTLS},
			registry: &types.Registry{
				Metadata: types.Metadata{Name: describeName},
				Spec:     types.RegistrySpec{Transport: types.TransportHTTP, TLS: &types.TLS{}},
			},
			wantContains: []string{labelName},
			wantAbsent:   []string{labelTLS},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			f := newDescribeFixture()
			f.store.On("FindByName", mock.Anything, describeName).Return(tt.registry, nil)
			var buf bytes.Buffer
			request := NewDescribeRegistryRequest(context.Background(), &buf)
			request.Name = describeName
			request.Fields = tt.fields

			// Act.
			err := f.uc.Execute(request)

			// Assert.
			require.NoError(t, err)
			out := buf.String()
			lastIndex := -1
			for _, want := range tt.wantContains {
				idx := strings.Index(out, want)
				require.GreaterOrEqualf(t, idx, 0, "output %q missing %q", out, want)
				assert.Greaterf(t, idx, lastIndex, "%q is out of order in %q", want, out)
				lastIndex = idx
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContainsf(t, out, absent, "output unexpectedly contains %q", absent)
			}
		})
	}
}

// TestDescribeRegistryFailure covers the not-found, io, and usage classifications.
func TestDescribeRegistryFailure(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// fields is the requested --fields selection.
		fields []string
		// found is the record FindByName returns.
		found *types.Registry
		// findErr is the error FindByName returns.
		findErr error
		// expectFind states whether FindByName is reached.
		expectFind bool
		// wantType is the expected error classification.
		wantType Type
	}{
		{
			name:       "unknown name is not found",
			expectFind: true,
			wantType:   TypeNotFound,
		},
		{
			name:       "store failure is io",
			found:      nil,
			findErr:    errors.New("disk gone"),
			expectFind: true,
			wantType:   TypeIO,
		},
		{
			name:     "unknown field is usage",
			fields:   []string{"bogus"},
			wantType: TypeUsage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			f := newDescribeFixture()
			if tt.expectFind {
				f.store.On("FindByName", mock.Anything, describeName).Return(tt.found, tt.findErr)
			}
			request := NewDescribeRegistryRequest(context.Background(), &bytes.Buffer{})
			request.Name = describeName
			request.Fields = tt.fields

			// Act.
			err := f.uc.Execute(request)

			// Assert.
			var ucErr *Error
			require.ErrorAs(t, err, &ucErr)
			assert.Equal(t, tt.wantType, ucErr.Type)
		})
	}
}

// TestDescribeRegistryRenderError surfaces a writer failure as an io error.
func TestDescribeRegistryRenderError(t *testing.T) {
	// Arrange.
	f := newDescribeFixture()
	f.store.On("FindByName", mock.Anything, describeName).Return(fullRegistry(), nil)
	request := NewDescribeRegistryRequest(context.Background(), failWriter{})
	request.Name = describeName

	// Act.
	err := f.uc.Execute(request)

	// Assert.
	var ucErr *Error
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, TypeIO, ucErr.Type)
}

// failWriter always fails, exercising the render error path.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
