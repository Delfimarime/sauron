package gherkin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorldRequiresBin(t *testing.T) {
	_, err := newWorld("", fakeEnv(nil))
	assert.Error(t, err)
}

func TestNewWorldPopulatesMaps(t *testing.T) {
	w, err := newWorld("/bin/echo", fakeEnv(nil))
	require.NoError(t, err)
	assert.Equal(t, "0.0.0", w.Environment["App"].(map[string]any)["Version"])
	assert.Equal(t, "/tmp/sauron-home", w.Variables["HomeDirectory"])
}

func TestWorldExecuteUsesHostRuntimeWhenNoSpecs(t *testing.T) {
	w, err := newWorld("/bin/echo", fakeEnv(nil))
	require.NoError(t, err)

	require.NoError(t, w.Execute(context.Background(), "hello"))
	require.NotNil(t, w.Last())
	assert.Equal(t, 0, w.Last().exitCode)
	assert.Equal(t, "hello\n", w.Last().output)

	require.NoError(t, w.Reset(context.Background()))
	assert.Nil(t, w.Last())
}
