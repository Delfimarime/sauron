package extension_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/delfimarime/sauron/pkg/sauron/extension"
)

// TestOptions asserts each functional option mutates exactly the Options
// field(s) it documents, by comparing the result against a fully specified
// expected struct. It is purely in-memory — no filesystem, no env.
func TestOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		option extension.Option
		want   extension.Options
	}{
		{
			name:   "WithURI sets URI",
			option: extension.WithURI("https://example.com/artifacts"),
			want:   extension.Options{URI: "https://example.com/artifacts"},
		},
		{
			name:   "WithRef sets Ref",
			option: extension.WithRef("v1.4.0"),
			want:   extension.Options{Ref: "v1.4.0"},
		},
		{
			name:   "WithTimeout sets Timeout",
			option: extension.WithTimeout(30 * time.Second),
			want:   extension.Options{Timeout: 30 * time.Second},
		},
		{
			name:   "WithBasicAuth sets Username and Password",
			option: extension.WithBasicAuth("alice", "s3cret"),
			want:   extension.Options{Username: "alice", Password: "s3cret"},
		},
		{
			name:   "WithSSHKey sets SSHKey",
			option: extension.WithSSHKey("/path/id_ed25519"),
			want:   extension.Options{SSHKey: "/path/id_ed25519"},
		},
		{
			name:   "WithSkipTLSVerify true toggles SkipTLSVerify",
			option: extension.WithSkipTLSVerify(true),
			want:   extension.Options{SkipTLSVerify: true},
		},
		{
			name:   "WithSkipTLSVerify false leaves SkipTLSVerify unset",
			option: extension.WithSkipTLSVerify(false),
			want:   extension.Options{SkipTLSVerify: false},
		},
		{
			name:   "WithCACert sets CACert",
			option: extension.WithCACert("/path/ca.pem"),
			want:   extension.Options{CACert: "/path/ca.pem"},
		},
		{
			name:   "WithClientCert sets ClientCert and ClientKey",
			option: extension.WithClientCert("/path/client.pem", "/path/client.key"),
			want:   extension.Options{ClientCert: "/path/client.pem", ClientKey: "/path/client.key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var o extension.Options

			// Act
			tt.option(&o)

			// Assert
			assert.Equal(t, tt.want, o)
		})
	}
}

// TestOptionsComposed asserts that applying several options together
// accumulates each field independently without interference.
func TestOptionsComposed(t *testing.T) {
	t.Parallel()

	// Arrange
	var o extension.Options
	options := []extension.Option{
		extension.WithURI("git@github.com:acme/artifacts.git"),
		extension.WithRef("main"),
		extension.WithTimeout(15 * time.Second),
		extension.WithBasicAuth("alice", "s3cret"),
		extension.WithSSHKey("/path/id_ed25519"),
		extension.WithSkipTLSVerify(true),
		extension.WithCACert("/path/ca.pem"),
		extension.WithClientCert("/path/client.pem", "/path/client.key"),
	}

	// Act
	for _, opt := range options {
		opt(&o)
	}

	// Assert
	assert.Equal(t, extension.Options{
		URI:           "git@github.com:acme/artifacts.git",
		Ref:           "main",
		Timeout:       15 * time.Second,
		Username:      "alice",
		Password:      "s3cret",
		SSHKey:        "/path/id_ed25519",
		SkipTLSVerify: true,
		CACert:        "/path/ca.pem",
		ClientCert:    "/path/client.pem",
		ClientKey:     "/path/client.key",
	}, o)
}
