package marketplace

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// x509PoolOf returns a certificate pool trusting the test server's certificate.
func x509PoolOf(t *testing.T, server *httptest.Server) *x509.CertPool {
	t.Helper()

	pool := x509.NewCertPool()
	require.NotNil(t, server.Certificate())
	pool.AddCert(server.Certificate())

	return pool
}

func TestClient_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection func(Client) ArtifactClient
		wantPath   string
	}{
		{
			name:       "skills list hits /skills",
			collection: func(c Client) ArtifactClient { return c.Skills() },
			wantPath:   "/skills",
		},
		{
			name:       "agents list hits /agents",
			collection: func(c Client) ArtifactClient { return c.Agents() },
			wantPath:   "/agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			var gotPath, gotQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = r.URL.RawQuery
				_, _ = w.Write([]byte(`{"items":[{"name":"writer","version":"1.2.0","size":42}]}`))
			}))
			defer server.Close()

			c, err := New(WithBaseURL(server.URL))
			require.NoError(t, err)

			// Act.
			list, listErr := tt.collection(c).List(context.Background(),
				WithSearch("wr"), WithSort("+name"), WithLimit(5), WithOffset(2))

			// Assert.
			require.NoError(t, listErr)
			assert.Equal(t, tt.wantPath, gotPath)
			assert.Contains(t, gotQuery, "q=wr")
			assert.Contains(t, gotQuery, "sort=%2Bname")
			assert.Contains(t, gotQuery, "limit=5")
			assert.Contains(t, gotQuery, "offset=2")
			require.Len(t, list.Items, 1)
			assert.Equal(t, "writer", list.Items[0].Name)
			require.NotNil(t, list.Items[0].Version)
			assert.Equal(t, "1.2.0", *list.Items[0].Version)
			require.NotNil(t, list.Items[0].Size)
			assert.Equal(t, int64(42), *list.Items[0].Size)
		})
	}
}

func TestClient_List_NullableFields(t *testing.T) {
	t.Parallel()

	// Arrange.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"name":"persona","version":null,"size":null}]}`))
	}))
	defer server.Close()

	c, err := New(WithBaseURL(server.URL))
	require.NoError(t, err)

	// Act.
	list, listErr := c.Agents().List(context.Background())

	// Assert: version/size stay nil.
	require.NoError(t, listErr)
	require.Len(t, list.Items, 1)
	assert.Nil(t, list.Items[0].Version)
	assert.Nil(t, list.Items[0].Size)
}

func TestClient_List_BasicAuth(t *testing.T) {
	t.Parallel()

	// Arrange.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "alice" || pass != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"status":401,"title":"Unauthorized"}`))
			return
		}
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	t.Run("valid credentials succeed", func(t *testing.T) {
		c, err := New(WithBaseURL(server.URL), WithBasicAuth("alice", "secret"))
		require.NoError(t, err)

		// Act.
		list, listErr := c.Skills().List(context.Background())

		// Assert.
		require.NoError(t, listErr)
		assert.Empty(t, list.Items)
	})

	t.Run("missing credentials are unauthorized", func(t *testing.T) {
		c, err := New(WithBaseURL(server.URL))
		require.NoError(t, err)

		// Act.
		_, listErr := c.Skills().List(context.Background())

		// Assert.
		require.Error(t, listErr)
		assert.True(t, IsUnauthorized(listErr))
	})
}

func TestClient_List_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     int
		body       string
		predicate  func(error) bool
		wantDetail string
	}{
		{
			name:       "not found carries problem detail",
			status:     http.StatusNotFound,
			body:       `{"type":"about:blank","title":"Not Found","detail":"no such collection"}`,
			predicate:  IsNotFound,
			wantDetail: "no such collection",
		},
		{
			name:      "forbidden",
			status:    http.StatusForbidden,
			body:      `{"title":"Forbidden"}`,
			predicate: IsForbidden,
		},
		{
			name:      "bad request",
			status:    http.StatusBadRequest,
			body:      `{"title":"Bad Request"}`,
			predicate: IsBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c, err := New(WithBaseURL(server.URL))
			require.NoError(t, err)

			// Act.
			_, listErr := c.Skills().List(context.Background())

			// Assert.
			require.Error(t, listErr)
			assert.True(t, tt.predicate(listErr))
			if tt.wantDetail != "" {
				assert.Contains(t, listErr.Error(), tt.wantDetail)
			}
		})
	}
}

func TestClient_List_TransportError(t *testing.T) {
	t.Parallel()

	// Arrange: a server that is closed before the request runs.
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	baseURL := server.URL
	server.Close()

	c, err := New(WithBaseURL(baseURL))
	require.NoError(t, err)

	// Act.
	_, listErr := c.Skills().List(context.Background())

	// Assert.
	require.Error(t, listErr)
	assert.ErrorIs(t, listErr, ErrTransport)
}

func TestClient_List_MalformedBody(t *testing.T) {
	t.Parallel()

	// Arrange.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	c, err := New(WithBaseURL(server.URL))
	require.NoError(t, err)

	// Act.
	_, listErr := c.Skills().List(context.Background())

	// Assert.
	require.Error(t, listErr)
	assert.ErrorIs(t, listErr, ErrTransport)
}

func TestNew_InvalidBaseURL(t *testing.T) {
	t.Parallel()

	// Act.
	_, err := New(WithBaseURL("://missing-scheme"))

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestNew_FullConfig(t *testing.T) {
	t.Parallel()

	// Arrange: a TLS server we trust via an explicit client TLS config.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"name":"writer"}]}`))
	}))
	defer server.Close()

	tlsConfig := &tls.Config{RootCAs: x509PoolOf(t, server), MinVersion: tls.VersionTLS12}

	c, err := New(
		WithBaseURL(server.URL),
		WithBasicAuth("u", "p"),
		WithTLSConfig(tlsConfig),
		WithTimeout(5*time.Second),
	)
	require.NoError(t, err)

	// Act.
	list, listErr := c.Agents().List(context.Background())

	// Assert.
	require.NoError(t, listErr)
	require.Len(t, list.Items, 1)
}

func TestArtifactClient_Content(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		collection func(Client) ArtifactClient
		wantPath   string
	}{
		{
			name:       "skills content hits /skills/writer/content",
			collection: func(c Client) ArtifactClient { return c.Skills() },
			wantPath:   "/skills/writer/content",
		},
		{
			name:       "agents content hits /agents/writer/content",
			collection: func(c Client) ArtifactClient { return c.Agents() },
			wantPath:   "/agents/writer/content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			fakeArchive := []byte("fake archive bytes")
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Artifact-Version", "abc123")
				_, _ = w.Write(fakeArchive)
			}))
			defer server.Close()

			c, err := New(WithBaseURL(server.URL))
			require.NoError(t, err)

			// Act.
			archive, version, contentErr := tt.collection(c).Content(context.Background(), "writer")

			// Assert.
			require.NoError(t, contentErr)
			assert.Equal(t, tt.wantPath, gotPath)
			assert.Equal(t, fakeArchive, archive)
			assert.Equal(t, "abc123", version)
		})
	}
}

func TestArtifactClient_Content_NoVersion(t *testing.T) {
	t.Parallel()

	// Arrange: server returns the archive without an Artifact-Version header.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("archive"))
	}))
	defer server.Close()

	c, err := New(WithBaseURL(server.URL))
	require.NoError(t, err)

	// Act.
	_, version, contentErr := c.Skills().Content(context.Background(), "writer")

	// Assert: version is empty, not an error.
	require.NoError(t, contentErr)
	assert.Empty(t, version)
}

func TestArtifactClient_Content_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		body      string
		predicate func(error) bool
	}{
		{
			name:      "404 is not-found",
			status:    http.StatusNotFound,
			body:      `{"status":404,"title":"Not Found","detail":"no such skill"}`,
			predicate: IsNotFound,
		},
		{
			name:      "401 is unauthorized",
			status:    http.StatusUnauthorized,
			body:      `{"status":401,"title":"Unauthorized"}`,
			predicate: IsUnauthorized,
		},
		{
			name:      "403 is forbidden",
			status:    http.StatusForbidden,
			body:      `{"status":403,"title":"Forbidden"}`,
			predicate: IsForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c, err := New(WithBaseURL(server.URL))
			require.NoError(t, err)

			// Act.
			_, _, contentErr := c.Skills().Content(context.Background(), "absent")

			// Assert.
			require.Error(t, contentErr)
			assert.True(t, tt.predicate(contentErr))
		})
	}
}

func TestArtifactClient_Content_TransportError(t *testing.T) {
	t.Parallel()

	// Arrange: close the server before issuing the request.
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	baseURL := server.URL
	server.Close()

	c, err := New(WithBaseURL(baseURL))
	require.NoError(t, err)

	// Act.
	_, _, contentErr := c.Skills().Content(context.Background(), "writer")

	// Assert.
	require.Error(t, contentErr)
	assert.ErrorIs(t, contentErr, ErrTransport)
}

func TestMockBasedArtifactClient_Content(t *testing.T) {
	t.Parallel()

	// Arrange.
	wantArchive := []byte("archive bytes")
	wantVersion := "v1.0.0"
	artifacts := &MockBasedArtifactClient{}
	artifacts.On("Content", mock.Anything, "writer").Return(wantArchive, wantVersion, nil)

	// Act.
	gotArchive, gotVersion, err := artifacts.Content(context.Background(), "writer")

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, wantArchive, gotArchive)
	assert.Equal(t, wantVersion, gotVersion)
	artifacts.AssertExpectations(t)
}

func TestAPIError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *APIError
		want string
	}{
		{name: "detail wins", err: &APIError{Status: 404, Detail: "missing"}, want: "registry responded 404: missing"},
		{name: "title fallback", err: &APIError{Status: 403, Title: "Forbidden"}, want: "registry responded 403: Forbidden"},
		{name: "status only", err: &APIError{Status: 500}, want: "registry responded 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestMockBasedClient(t *testing.T) {
	t.Parallel()

	// Arrange.
	want := &ArtifactList{Items: []ArtifactSummary{{Name: "writer"}}}
	artifacts := &MockBasedArtifactClient{}
	artifacts.On("List", mock.Anything, mock.Anything).Return(want, nil)

	client := &MockBasedClient{}
	client.On("Skills").Return(artifacts)
	client.On("Agents").Return(ArtifactClient(nil))

	// Act.
	got, err := client.Skills().List(context.Background(), WithLimit(1))

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Nil(t, client.Agents())
	client.AssertExpectations(t)
	artifacts.AssertExpectations(t)
}

func TestPredicates_NonAPIError(t *testing.T) {
	t.Parallel()

	// Assert: a plain error is none of the API conditions.
	assert.False(t, IsNotFound(ErrTransport))
	assert.False(t, IsUnauthorized(ErrTransport))
	assert.False(t, IsForbidden(ErrTransport))
	assert.False(t, IsBadRequest(ErrTransport))
}
