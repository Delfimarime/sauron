package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSynthCompose(t *testing.T) {
	specs := []ContainerSpec{
		{Service: "git", Image: "gitea/gitea:1", Ports: []string{"3000/tcp"}, Env: map[string]string{"K": "V"}},
		{Service: "http", Image: "nginx:alpine"},
	}

	doc, err := GenerateDockerComposeFile(specs)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, yaml.Unmarshal(doc, &got))

	services, ok := got["services"].(map[string]any)
	require.True(t, ok, "document has a services map")
	assert.Len(t, services, 2)

	git, ok := services["git"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "gitea/gitea:1", git["image"])
	assert.Equal(t, []any{"3000/tcp"}, git["ports"])
	assert.Equal(t, map[string]any{"K": "V"}, git["environment"])

	http, ok := services["http"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "nginx:alpine", http["image"])
	assert.NotContains(t, http, "ports")
	assert.NotContains(t, http, "environment")
}

func TestSynthComposeErrors(t *testing.T) {
	tests := map[string]struct {
		specs []ContainerSpec
	}{
		"no specs":        {specs: nil},
		"missing image":   {specs: []ContainerSpec{{Service: "x"}}},
		"missing service": {specs: []ContainerSpec{{Image: "x"}}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := GenerateDockerComposeFile(tc.specs)
			assert.Error(t, err)
		})
	}
}
