package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// descriptor-test literals, named to satisfy goconst across the package.
const (
	viewLabelTransport = "transport"
	viewLabelURI       = "uri"
	viewLabelAuth      = "auth"
	labelUsername  = "username"
	valGit         = "git"
	valUserRef     = "${env:ACME_USER}"
)

// TestDescriptorRender exercises the rendering rules: aligned leaf values, the
// nested section block, and the no-output-for-zero-fields rule.
func TestDescriptorRender(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// descriptor is the value under test.
		descriptor Descriptor
		// want is the exact expected output.
		want string
	}{
		{
			name:       "no fields produce no output",
			descriptor: Descriptor{},
			want:       "",
		},
		{
			name: "leaf values align to the widest label",
			descriptor: Descriptor{Fields: []Field{
				{Label: viewColName, Value: rowAcme},
				{Label: viewLabelTransport, Value: valGit},
				{Label: viewLabelURI, Value: acmeURI},
			}},
			want: "name:       acme\n" +
				"transport:  git\n" +
				"uri:        git@github.com:acme/artifacts.git\n",
		},
		{
			name: "a section renders its children indented and aligned",
			descriptor: Descriptor{Fields: []Field{
				{Label: viewColName, Value: rowAcme},
				{Label: viewLabelTransport, Value: valGit},
				{Label: viewLabelAuth, Children: []Field{
					{Label: labelUsername, Value: valUserRef},
					{Label: "password", Value: "${env:ACME_TOKEN}"},
				}},
				{Label: fieldTimeout, Value: "30s"},
			}},
			want: "name:       acme\n" +
				"transport:  git\n" +
				"auth:\n" +
				"  username: ${env:ACME_USER}\n" +
				"  password: ${env:ACME_TOKEN}\n" +
				"timeout:    30s\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := tt.descriptor.Render(&buf)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestDescriptorRenderWriteError surfaces a writer failure rather than swallowing
// it, on both a leaf line and a section header line.
func TestDescriptorRenderWriteError(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// descriptor is the value under test.
		descriptor Descriptor
		// writeAfter is the number of successful writes before the failure.
		writeAfter int
	}{
		{
			name:       "leaf line write fails",
			descriptor: Descriptor{Fields: []Field{{Label: "name", Value: "acme"}}},
			writeAfter: 0,
		},
		{
			name: "section header write fails",
			descriptor: Descriptor{Fields: []Field{
				{Label: "auth", Children: []Field{{Label: labelUsername, Value: "u"}}},
			}},
			writeAfter: 0,
		},
		{
			name: "section child write fails",
			descriptor: Descriptor{Fields: []Field{
				{Label: "auth", Children: []Field{{Label: labelUsername, Value: "u"}}},
			}},
			writeAfter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := tt.descriptor.Render(&failingWriter{writeAfter: tt.writeAfter})

			// Assert.
			require.Error(t, err)
		})
	}
}
