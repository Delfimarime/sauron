package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// headingSkills is the skills group heading reused across the cases.
const headingSkills = "skills"

// TestRenderGroups pins the heading-and-entries layout and the empty-group skip.
func TestRenderGroups(t *testing.T) {
	tests := []struct {
		// name states the case intent.
		name string
		// groups is the input.
		groups []Group
		// want is the exact expected output.
		want string
	}{
		{
			name:   "no groups produce no output",
			groups: nil,
			want:   "",
		},
		{
			name: "only non-empty groups render",
			groups: []Group{
				{Heading: headingSkills, Items: []string{"sauron-acme-go-style"}},
				{Heading: "agents", Items: []string{"sauron-acme-code-reviewer"}},
				{Heading: "personas", Items: nil},
			},
			want: "skills:\n  - sauron-acme-go-style\n" +
				"agents:\n  - sauron-acme-code-reviewer\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			var buf bytes.Buffer

			// Act.
			err := RenderGroups(&buf, tt.groups)

			// Assert.
			require.NoError(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

// TestRenderGroupsWriteError surfaces a writer failure on both the heading line
// and an item line.
func TestRenderGroupsWriteError(t *testing.T) {
	groups := []Group{{Heading: headingSkills, Items: []string{"a", "b"}}}

	for _, writeAfter := range []int{0, 1} {
		err := RenderGroups(&failingWriter{writeAfter: writeAfter}, groups)
		require.Error(t, err)
	}
}
