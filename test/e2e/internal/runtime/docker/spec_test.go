//go:build unit

package docker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGenerateDockerComposeFile(t *testing.T) {
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

func TestGenerateDockerComposeFileRendersEntrypointAndVolumes(t *testing.T) {
	specs := []ContainerSpec{{
		Service:    "main",
		Image:      "alpine:3",
		Entrypoint: "tail -f /dev/null",
		Mount:      []FileSpec{{SourceFile: "/host/sauron", Path: "/opt/bin/sauron"}},
	}}

	doc, err := GenerateDockerComposeFile(specs)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, yaml.Unmarshal(doc, &got))
	main := got["services"].(map[string]any)["main"].(map[string]any)

	assert.Equal(t, []any{"tail", "-f", "/dev/null"}, main["entrypoint"])
	assert.Equal(t, []any{"/host/sauron:/opt/bin/sauron"}, main["volumes"])
}

func TestGenerateDockerComposeFileErrors(t *testing.T) {
	tests := map[string]struct {
		specs []ContainerSpec
	}{
		"no specs":            {specs: nil},
		"missing image":       {specs: []ContainerSpec{{Service: "x"}}},
		"missing service":     {specs: []ContainerSpec{{Image: "x"}}},
		"mount without source": {specs: []ContainerSpec{{Service: "x", Image: "img", Mount: []FileSpec{{Path: "/p", Content: []byte("hi")}}}}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := GenerateDockerComposeFile(tc.specs)
			assert.Error(t, err)
		})
	}
}

func TestMaterializeContent(t *testing.T) {
	written := map[string][]byte{}
	writeFile := func(name string, data []byte, _ os.FileMode) error {
		written[name] = data
		return nil
	}

	specs := []ContainerSpec{{
		Service: "main",
		Image:   "alpine:3",
		Mount: []FileSpec{
			{Path: "/etc/config.yml", Content: []byte("a: 1")},      // materialized
			{Path: "/opt/bin/sauron", SourceFile: "/host/sauron"},   // left as-is
		},
	}}

	out, err := materializeContent(writeFile, "/tmp/work", specs)
	require.NoError(t, err)

	materialized := out[0].Mount[0]
	assert.Equal(t, "/tmp/work/main-0-config.yml", materialized.SourceFile)
	assert.Equal(t, []byte("a: 1"), written["/tmp/work/main-0-config.yml"])

	assert.Equal(t, "/host/sauron", out[0].Mount[1].SourceFile, "existing SourceFile is left untouched")
	assert.Len(t, written, 1, "only Content mounts are written")
}
