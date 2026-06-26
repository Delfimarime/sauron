package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// TestNewValidatorRegistersKinds loads every embedded schema, including Registry.
func TestNewValidatorRegistersKinds(t *testing.T) {
	// Arrange + Act.
	v, err := newValidator()

	// Assert.
	require.NoError(t, err)
	require.NotNil(t, v)
	assert.Contains(t, v.byKind, types.KindRegistry)
}

// TestValidatorValidate exercises accept, reject, and unknown-kind paths.
func TestValidatorValidate(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// kind is the document kind to validate against.
		kind string
		// doc is the YAML document under test.
		doc string
		// wantErr asserts validation fails.
		wantErr bool
	}{
		{
			name: "accepts a valid registry",
			kind: types.KindRegistry,
			doc: `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: git
  source: https://example.com/acme.git
`,
		},
		{
			name: "rejects an invalid transport",
			kind: types.KindRegistry,
			doc: `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: ftp
  source: https://example.com/acme.git
`,
			wantErr: true,
		},
		{
			name: "rejects a missing required field",
			kind: types.KindRegistry,
			doc: `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: git
`,
			wantErr: true,
		},
		{
			name:    "rejects an unknown kind",
			kind:    "Nonexistent",
			doc:     "metadata:\n  name: x\n",
			wantErr: true,
		},
	}

	v, err := newValidator()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var node yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(tt.doc), &node))

			// Act.
			err := v.validate(tt.kind, &node)

			// Assert.
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
