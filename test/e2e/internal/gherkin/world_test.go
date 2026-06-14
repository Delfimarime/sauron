//go:build unit

package gherkin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRuntime is a runtime.Runtime stub: it records nothing and returns a canned
// result, so World can be tested without a host process or Docker.
type fakeRuntime struct {
	code int
	out  string
	err  error
}

func (fakeRuntime) IsReadOnly() bool            { return true }
func (fakeRuntime) Start(context.Context) error { return nil }
func (fakeRuntime) Stop(context.Context) error  { return nil }
func (f fakeRuntime) Execute(context.Context, ...string) (int, string, error) {
	return f.code, f.out, f.err
}

func TestNewWorldPopulatesMaps(t *testing.T) {
	w, err := newWorld(fakeRuntime{}, fakeEnv(nil))
	require.NoError(t, err)
	assert.Equal(t, "0.0.0", w.Environment["App"].(map[string]any)["Version"])
	assert.Equal(t, "/tmp/sauron-home", w.Variables["HomeDirectory"])
}

func TestNewWorldRequiresEnv(t *testing.T) {
	_, err := newWorld(fakeRuntime{}, fakeEnv(map[string]string{"SAURON_APP_VERSION": ""}))
	assert.Error(t, err)
}

func TestWorldExecuteRecordsResult(t *testing.T) {
	w, err := newWorld(fakeRuntime{code: 0, out: "hello\n"}, fakeEnv(nil))
	require.NoError(t, err)

	require.NoError(t, w.Execute(context.Background(), "x"))
	require.NotNil(t, w.Last())
	assert.Equal(t, 0, w.Last().exitCode)
	assert.Equal(t, "hello\n", w.Last().output)
}
