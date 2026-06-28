package registry

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// writerName is the fixture entry name returned by the stub registry server.
const writerName = "writer"

func TestRESTFactory_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []extension.Option
		wantErr bool
	}{
		{
			name: "auth and tls are accepted",
			opts: []extension.Option{
				extension.WithURI("https://registry.example"),
				extension.WithBasicAuth("u", "p"),
				extension.WithSkipTLSVerify(true),
			},
		},
		{
			name:    "reference is rejected",
			opts:    []extension.Option{extension.WithURI("https://registry.example"), extension.WithRef("main")},
			wantErr: true,
		},
		{
			name:    "ssh key is rejected",
			opts:    []extension.Option{extension.WithURI("https://registry.example"), extension.WithSSHKey("/id")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			err := newRESTFactory().Validate(tt.opts...)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, api.ErrUsage)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRESTFactory_Open_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		uri         string
		wantPath    string
		wantNames   []string
		wantVersion string
	}{
		{
			name:        "skills collection maps to /skills",
			uri:         "skills",
			wantPath:    "/skills",
			wantNames:   []string{writerName},
			wantVersion: "1.2.0",
		},
		{
			name:      "agents collection maps to /agents",
			uri:       "agents",
			wantPath:  "/agents",
			wantNames: []string{writerName},
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

			fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
			require.NoError(t, err)

			// Act.
			files, listErr := fs.List(context.Background(), tt.uri,
				source.WithLimit(1), source.WithSearch("wr"), source.WithOffset(2), source.WithSort("name"))

			// Assert.
			require.NoError(t, listErr)
			gotValues, parseErr := url.ParseQuery(gotQuery)
			require.NoError(t, parseErr)
			assert.Equal(t, tt.wantPath, gotPath)
			assert.Contains(t, gotQuery, "limit=1")
			assert.Contains(t, gotQuery, "q=wr")
			assert.Contains(t, gotQuery, "offset=2")
			assert.Equal(t, "+name", gotValues.Get("sort"))
			require.Len(t, files, len(tt.wantNames))
			assert.Equal(t, tt.wantNames[0], files[0].Name())
			assert.Equal(t, int64(42), files[0].Size())
			assert.False(t, files[0].IsDirectory())
			if tt.wantVersion != "" {
				assert.Equal(t, tt.wantVersion, files[0].Version())
			}
		})
	}
}

func TestRESTFactory_Open_List_Order(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     []source.Option
		wantSort string
	}{
		{
			name:     "ascending order signs the sort directive with +",
			opts:     []source.Option{source.WithSort("name"), source.WithOrder("asc")},
			wantSort: "+name",
		},
		{
			name:     "descending order signs the sort directive with -",
			opts:     []source.Option{source.WithSort("name"), source.WithOrder("desc")},
			wantSort: "-name",
		},
		{
			name:     "unset order defaults to ascending",
			opts:     []source.Option{source.WithSort("name")},
			wantSort: "+name",
		},
		{
			name:     "order without sort sends no sort directive",
			opts:     []source.Option{source.WithOrder("desc")},
			wantSort: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			var gotQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.RawQuery
				_, _ = w.Write([]byte(`{"items":[]}`))
			}))
			defer server.Close()

			fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
			require.NoError(t, err)

			// Act.
			_, listErr := fs.List(context.Background(), "skills", tt.opts...)

			// Assert.
			require.NoError(t, listErr)
			gotValues, parseErr := url.ParseQuery(gotQuery)
			require.NoError(t, parseErr)
			assert.Equal(t, tt.wantSort, gotValues.Get("sort"))
		})
	}
}

func TestRESTFactory_Open_List_BasicAuth(t *testing.T) {
	t.Parallel()

	// Arrange.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "alice" || pass != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	fs, err := newRESTFactory().Open(context.Background(),
		extension.WithURI(server.URL), extension.WithBasicAuth("alice", "secret"))
	require.NoError(t, err)

	// Act.
	files, listErr := fs.List(context.Background(), "skills", source.WithLimit(1))

	// Assert.
	require.NoError(t, listErr)
	assert.Empty(t, files)
}

func TestRESTFactory_Open_List_Errors(t *testing.T) {
	t.Parallel()

	t.Run("non-2xx is a runtime error", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
		require.NoError(t, err)

		// Act.
		_, listErr := fs.List(context.Background(), "skills", source.WithLimit(1))

		// Assert.
		require.Error(t, listErr)
		assert.ErrorIs(t, listErr, api.ErrRuntime)
	})

	t.Run("malformed body is a runtime error", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
		require.NoError(t, err)

		// Act.
		_, listErr := fs.List(context.Background(), "skills", source.WithLimit(1))

		// Assert.
		require.Error(t, listErr)
		assert.ErrorIs(t, listErr, api.ErrRuntime)
	})

	t.Run("unreachable host is a runtime error", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := server.URL
		server.Close()

		fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(url))
		require.NoError(t, err)

		// Act.
		_, listErr := fs.List(context.Background(), "skills", source.WithLimit(1))

		// Assert.
		require.Error(t, listErr)
		assert.ErrorIs(t, listErr, api.ErrRuntime)
	})

	t.Run("unknown collection is a usage error", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		fs, err := newRESTFactory().Open(context.Background(), extension.WithURI("https://registry.example"))
		require.NoError(t, err)

		// Act.
		_, listErr := fs.List(context.Background(), ".widgets")

		// Assert.
		require.Error(t, listErr)
		assert.ErrorIs(t, listErr, api.ErrUsage)
	})

	t.Run("unauthorized is a usage error", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
		require.NoError(t, err)

		// Act.
		_, listErr := fs.List(context.Background(), "skills")

		// Assert.
		require.Error(t, listErr)
		assert.ErrorIs(t, listErr, api.ErrUsage)
	})

	t.Run("invalid options surface on open", func(t *testing.T) {
		t.Parallel()

		// Act.
		_, err := newRESTFactory().Open(context.Background(),
			extension.WithURI("https://registry.example"), extension.WithRef("main"))

		// Assert.
		require.Error(t, err)
		assert.ErrorIs(t, err, api.ErrUsage)
	})

	t.Run("invalid ca certificate is a usage error", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		bad := writeTempFile(t, "ca.pem", "not a certificate")

		// Act.
		_, err := newRESTFactory().Open(context.Background(),
			extension.WithURI("https://registry.example"), extension.WithCACert(bad))

		// Assert.
		require.Error(t, err)
		assert.ErrorIs(t, err, api.ErrUsage)
	})
}

func TestRESTFactory_Open_TLS(t *testing.T) {
	t.Parallel()

	// Arrange: a TLS server whose CA we trust explicitly.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"name":"writer","version":"1","size":1}]}`))
	}))
	defer server.Close()

	caPath := writeTempFile(t, "ca.pem", string(serverCAPEM(t, server)))

	fs, err := newRESTFactory().Open(context.Background(),
		extension.WithURI(server.URL), extension.WithCACert(caPath))
	require.NoError(t, err)

	// Act.
	files, listErr := fs.List(context.Background(), "skills", source.WithLimit(1))

	// Assert.
	require.NoError(t, listErr)
	require.Len(t, files, 1)
}

func TestRESTFactory_Open_SkipTLSVerify(t *testing.T) {
	t.Parallel()

	// Arrange: a TLS server with an untrusted certificate.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	// Without skipping verification the handshake fails.
	strict, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
	require.NoError(t, err)
	_, strictErr := strict.List(context.Background(), "skills", source.WithLimit(1))
	require.Error(t, strictErr)
	assert.ErrorIs(t, strictErr, api.ErrRuntime)

	// Skipping verification succeeds.
	lax, err := newRESTFactory().Open(context.Background(),
		extension.WithURI(server.URL), extension.WithSkipTLSVerify(true))
	require.NoError(t, err)

	// Act.
	files, listErr := lax.List(context.Background(), "skills", source.WithLimit(1))

	// Assert.
	require.NoError(t, listErr)
	assert.Empty(t, files)
}

func TestRESTFactory_Open_ClientCert(t *testing.T) {
	t.Parallel()

	t.Run("valid client certificate is loaded", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		certPath, keyPath := writeClientCert(t)

		// Act.
		_, err := newRESTFactory().Open(context.Background(),
			extension.WithURI("https://registry.example"),
			extension.WithClientCert(certPath, keyPath))

		// Assert.
		require.NoError(t, err)
	})

	t.Run("invalid client certificate is a usage error", func(t *testing.T) {
		t.Parallel()

		// Arrange.
		cert := writeTempFile(t, "cert.pem", "not a certificate")
		key := writeTempFile(t, "key.pem", "not a key")

		// Act.
		_, err := newRESTFactory().Open(context.Background(),
			extension.WithURI("https://registry.example"),
			extension.WithClientCert(cert, key))

		// Assert.
		require.Error(t, err)
		assert.ErrorIs(t, err, api.ErrUsage)
	})
}

func TestRESTFactory_Open_InvalidURI(t *testing.T) {
	t.Parallel()

	// Act.
	_, err := newRESTFactory().Open(context.Background(), extension.WithURI("://missing-scheme"))

	// Assert.
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrUsage)
}

func TestRESTFile_ReadIsNotImplemented(t *testing.T) {
	t.Parallel()

	// Arrange.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"name":"writer","version":"1","size":1}]}`))
	}))
	defer server.Close()

	fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
	require.NoError(t, err)
	files, listErr := fs.List(context.Background(), "skills")
	require.NoError(t, listErr)
	require.Len(t, files, 1)

	// Act.
	_, readErr := files[0].Read(context.Background())

	// Assert.
	assert.ErrorIs(t, readErr, source.ErrNotImplemented)
}

func TestRESTFactory_Open_DescribeAndGetNotImplemented(t *testing.T) {
	t.Parallel()

	// Arrange.
	fs, err := newRESTFactory().Open(context.Background(), extension.WithURI("https://registry.example"))
	require.NoError(t, err)

	// Act.
	_, describeErr := fs.Describe(context.Background(), "skills")
	_, getErr := fs.Get(context.Background(), "skills/writer")

	// Assert.
	assert.ErrorIs(t, describeErr, source.ErrNotImplemented)
	assert.ErrorIs(t, getErr, source.ErrNotImplemented)
}

func TestRESTFactory_Open_Fetch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		uri         string
		prefix      string // prefix used inside the archive entries
		wantVersion string
		wantFiles   map[string]string
	}{
		{
			name:        "strips kind/name prefix from archive entries",
			uri:         "skills/writer",
			prefix:      "skills/writer/",
			wantVersion: "abc123",
			wantFiles: map[string]string{
				skillMD:     "# writer skill",
				"lib/util.go": "package lib",
			},
		},
		{
			name:        "strips name-only prefix from archive entries",
			uri:         "skills/writer",
			prefix:      "writer/",
			wantVersion: "abc123",
			wantFiles: map[string]string{
				skillMD: "# writer skill",
			},
		},
		{
			name:        "agents endpoint is called for agents uri",
			uri:         "agents/reviewer",
			prefix:      "reviewer/",
			wantVersion: "def456",
			wantFiles: map[string]string{
				"AGENT.md": "# reviewer agent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: serve a real tar.gz archive with the chosen prefix.
			archive := makeGZipTar(t, tt.prefix, tt.wantFiles)
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Artifact-Version", tt.wantVersion)
				_, _ = w.Write(archive)
			}))
			defer server.Close()

			fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
			require.NoError(t, err)

			// Act.
			files, fetchErr := fs.Fetch(context.Background(), tt.uri)

			// Assert: correct endpoint, expected file count, relative names, content and version.
			require.NoError(t, fetchErr)
			assert.Equal(t, "/"+tt.uri+"/content", gotPath)
			require.Len(t, files, len(tt.wantFiles))

			byName := make(map[string]source.File, len(files))
			for _, f := range files {
				byName[f.Name()] = f
			}
			for wantName, wantBody := range tt.wantFiles {
				f, ok := byName[wantName]
				require.True(t, ok, "expected file %q in result", wantName)
				assert.Equal(t, tt.wantVersion, f.Version())
				assert.False(t, f.IsDirectory())
				assert.Equal(t, int64(len(wantBody)), f.Size())

				r, readErr := f.Read(context.Background())
				require.NoError(t, readErr)
				body, _ := io.ReadAll(r)
				require.NoError(t, r.Close())
				assert.Equal(t, wantBody, string(body))
			}
		})
	}
}

func TestRESTFactory_Open_Fetch_NoVersion(t *testing.T) {
	t.Parallel()

	// Arrange: server returns no Artifact-Version header.
	archive := makeGZipTar(t, "writer/", map[string]string{skillMD: "content"})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
	require.NoError(t, err)

	// Act.
	files, fetchErr := fs.Fetch(context.Background(), "skills/writer")

	// Assert: files are returned with empty version (not an error).
	require.NoError(t, fetchErr)
	require.Len(t, files, 1)
	assert.Empty(t, files[0].Version())
}

func TestRESTFactory_Open_Fetch_InvalidURI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		uri  string
	}{
		{name: "kind without name", uri: "skills"},
		{name: "empty name after slash", uri: "skills/"},
		{name: "empty string", uri: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			fs, err := newRESTFactory().Open(context.Background(), extension.WithURI("https://registry.example"))
			require.NoError(t, err)

			// Act.
			_, fetchErr := fs.Fetch(context.Background(), tt.uri)

			// Assert.
			require.Error(t, fetchErr)
			assert.ErrorIs(t, fetchErr, api.ErrUsage)
		})
	}
}

func TestRESTFactory_Open_Fetch_UnknownKind(t *testing.T) {
	t.Parallel()

	// Arrange.
	fs, err := newRESTFactory().Open(context.Background(), extension.WithURI("https://registry.example"))
	require.NoError(t, err)

	// Act.
	_, fetchErr := fs.Fetch(context.Background(), "widgets/foo")

	// Assert.
	require.Error(t, fetchErr)
	assert.ErrorIs(t, fetchErr, api.ErrUsage)
}

func TestRESTFactory_Open_Fetch_ServerError(t *testing.T) {
	t.Parallel()

	// Arrange: server returns 500.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
	require.NoError(t, err)

	// Act.
	_, fetchErr := fs.Fetch(context.Background(), "skills/writer")

	// Assert.
	require.Error(t, fetchErr)
	assert.ErrorIs(t, fetchErr, api.ErrRuntime)
}

func TestRESTFactory_Open_Fetch_NotFound(t *testing.T) {
	t.Parallel()

	// Arrange: server returns 404.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"title":"Not Found"}`))
	}))
	defer server.Close()

	fs, err := newRESTFactory().Open(context.Background(), extension.WithURI(server.URL))
	require.NoError(t, err)

	// Act.
	_, fetchErr := fs.Fetch(context.Background(), "skills/absent")

	// Assert: 404 is surfaced as a runtime error (not offered → caller decides).
	require.Error(t, fetchErr)
	assert.ErrorIs(t, fetchErr, api.ErrRuntime)
}
