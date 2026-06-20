package registry

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// names projects the file names from a listing.
func names(files []source.File) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.Name())
	}
	return out
}

// writeTempFile writes content into an isolated temporary file and returns its
// path.
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

// serverCAPEM returns the PEM-encoded certificate the test TLS server presents.
func serverCAPEM(t *testing.T, server *httptest.Server) []byte {
	t.Helper()

	cert := server.Certificate()
	require.NotNil(t, cert)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

// writeClientCert generates a self-signed certificate/key pair and writes them
// to isolated temporary files, returning their paths.
func writeClientCert(t *testing.T) (certPath, keyPath string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "sauron-test-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return writeTempFile(t, "client.crt", string(certPEM)), writeTempFile(t, "client.key", string(keyPEM))
}
