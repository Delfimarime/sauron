package gherkin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeEnv(overrides map[string]string) func(string) string {
	base := map[string]string{
		"SAURON_APP_NAME":    "sauron",
		"SAURON_APP_VERSION": "0.0.0",
		"SAURON_APP_TAG":     "SNAPSHOT",
		"SAURON_GIT_HASH":    "abc1234",
		"SAURON_HOME":        "/tmp/sauron-home",
	}
	for k, v := range overrides {
		if v == "" {
			delete(base, k)
		} else {
			base[k] = v
		}
	}
	return func(key string) string { return base[key] }
}

func TestNewEnvironment(t *testing.T) {
	env, err := newEnvironment(fakeEnv(nil))
	require.NoError(t, err)

	app := env["App"].(map[string]any)
	assert.Equal(t, "sauron", app["Name"])
	assert.Equal(t, "0.0.0", app["Version"])
	assert.Equal(t, "SNAPSHOT", app["Tag"])
	assert.Equal(t, "abc1234", env["GitHash"])
}

func TestNewEnvironmentMissingRequiredErrors(t *testing.T) {
	for _, key := range requiredEnvVars {
		t.Run(key, func(t *testing.T) {
			_, err := newEnvironment(fakeEnv(map[string]string{key: ""}))
			assert.Error(t, err)
		})
	}
}

func TestNewVariables(t *testing.T) {
	assert.Equal(t, "/tmp/sauron-home",
		newVariables(fakeEnv(nil))["HomeDirectory"])
	assert.Equal(t, "~/.sauron",
		newVariables(fakeEnv(map[string]string{"SAURON_HOME": ""}))["HomeDirectory"])
}
