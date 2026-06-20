package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/delfimarime/sauron/pkg/sauron/extension"
)

func TestResolve(t *testing.T) {
	t.Parallel()

	// Act.
	options := Resolve([]extension.Option{
		extension.WithURI("/srv"),
		extension.WithRef("main"),
	})

	// Assert.
	assert.Equal(t, "/srv", options.URI)
	assert.Equal(t, "main", options.Ref)
}

func TestHasAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []extension.Option
		want bool
	}{
		{name: "no credentials", want: false},
		{name: "username only", opts: []extension.Option{extension.WithBasicAuth("u", "")}, want: true},
		{name: "password only", opts: []extension.Option{extension.WithBasicAuth("", "p")}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, HasAuth(Resolve(tt.opts)))
		})
	}
}

func TestHasTLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []extension.Option
		want bool
	}{
		{name: "no tls", want: false},
		{name: "skip verify", opts: []extension.Option{extension.WithSkipTLSVerify(true)}, want: true},
		{name: "ca cert", opts: []extension.Option{extension.WithCACert("/ca.pem")}, want: true},
		{name: "client cert", opts: []extension.Option{extension.WithClientCert("/c.pem", "/k.pem")}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, HasTLS(Resolve(tt.opts)))
		})
	}
}
