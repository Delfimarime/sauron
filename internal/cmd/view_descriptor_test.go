package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// descriptor-view-test literals, named to satisfy goconst across the package.
const (
	dLabelTransport = "transport"
	dLabelURI       = "uri"
	dLabelAuth      = "auth"
	dLabelUsername  = "username"
	dValGit         = "git"
	dValUserRef     = "${env:ACME_USER}"
)

// TestDescriptorRender exercises the rendering rules: aligned leaf values, the
// nested section block, and the no-output-for-zero-fields rule.
func TestDescriptorRender(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// view is the value under test.
		view descriptor
		// want is the exact expected output.
		want string
	}{
		{
			name: "no fields produce no output",
			view: descriptor{},
			want: "",
		},
		{
			name: "leaf values align to the widest label",
			view: descriptor{Fields: []descriptorField{
				{Label: tblColName, Value: tblRowAcme},
				{Label: dLabelTransport, Value: dValGit},
				{Label: dLabelURI, Value: vGitURI},
			}},
			want: "name:       acme\n" +
				"transport:  git\n" +
				"uri:        git@github.com:acme/artifacts.git\n",
		},
		{
			name: "a section renders its children indented and aligned",
			view: descriptor{Fields: []descriptorField{
				{Label: tblColName, Value: tblRowAcme},
				{Label: dLabelTransport, Value: dValGit},
				{Label: dLabelAuth, Children: []descriptorField{
					{Label: dLabelUsername, Value: dValUserRef},
					{Label: "password", Value: "${env:ACME_TOKEN}"},
				}},
				{Label: describeFieldTimeout, Value: "30s"},
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
			err := tt.view.render(&buf)

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
		// view is the value under test.
		view descriptor
		// writeAfter is the number of successful writes before the failure.
		writeAfter int
	}{
		{
			name:       "leaf line write fails",
			view:       descriptor{Fields: []descriptorField{{Label: "name", Value: "acme"}}},
			writeAfter: 0,
		},
		{
			name: "section header write fails",
			view: descriptor{Fields: []descriptorField{
				{Label: "auth", Children: []descriptorField{{Label: dLabelUsername, Value: "u"}}},
			}},
			writeAfter: 0,
		},
		{
			name: "section child write fails",
			view: descriptor{Fields: []descriptorField{
				{Label: "auth", Children: []descriptorField{{Label: dLabelUsername, Value: "u"}}},
			}},
			writeAfter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			err := tt.view.render(&failingWriter{writeAfter: tt.writeAfter})

			// Assert.
			require.Error(t, err)
		})
	}
}
