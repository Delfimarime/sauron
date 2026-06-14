//go:build unit

package gherkin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleBanner = "sauron v0.0.0-SNAPSHOT\nHash abc1234\nHome: /tmp/h\n"

// fakeRuntime is a runtime.Runtime stub returning a canned result and recording
// the args it was called with, so basicController can be tested without a host
// process or Docker.
type fakeRuntime struct {
	code int
	out  string
	err  error
	args []string
}

func (*fakeRuntime) IsReadOnly() bool                          { return false }
func (*fakeRuntime) Start(context.Context) error               { return nil }
func (*fakeRuntime) Stop(context.Context) error                { return nil }
func (*fakeRuntime) CopyTo(context.Context, string, []byte) error { return nil }
func (f *fakeRuntime) Execute(_ context.Context, args ...string) (int, string, error) {
	f.args = args
	return f.code, f.out, f.err
}

func TestBasicControllerIsVersion(t *testing.T) {
	rt := &fakeRuntime{out: sampleBanner}
	b := &basicController{rt: rt}

	require.NoError(t, b.IsVersion(context.Background(), "0.0.0-SNAPSHOT"))
	assert.Equal(t, []string{"sauron", "--version"}, rt.args, "runs the binary under test")

	assert.Error(t, b.IsVersion(context.Background(), "9.9.9"), "mismatched version fails")
}

func TestBasicControllerIsVersionNonZeroExit(t *testing.T) {
	b := &basicController{rt: &fakeRuntime{code: 1, out: "boom"}}
	assert.Error(t, b.IsVersion(context.Background(), "0.0.0-SNAPSHOT"))
}

func TestBasicControllerIsVersionExecError(t *testing.T) {
	b := &basicController{rt: &fakeRuntime{err: errors.New("no runtime")}}
	assert.Error(t, b.IsVersion(context.Background(), "0.0.0-SNAPSHOT"))
}

func TestBasicControllerIsHomeDirectory(t *testing.T) {
	b := &basicController{rt: &fakeRuntime{out: sampleBanner}}

	require.NoError(t, b.IsHomeDirectory(context.Background(), "/tmp/h"))
	assert.Error(t, b.IsHomeDirectory(context.Background(), "/somewhere/else"))
}

func TestBasicControllerParseVersion(t *testing.T) {
	b := &basicController{}

	v, err := b.parseVersion(sampleBanner)
	require.NoError(t, err)
	assert.Equal(t, "0.0.0-SNAPSHOT", v)

	_, err = b.parseVersion("no-version-marker\n")
	assert.Error(t, err)
}

func TestBasicControllerParseHomeDirectory(t *testing.T) {
	b := &basicController{}

	h, err := b.parseHomeDirectory(sampleBanner)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/h", h)

	_, err = b.parseHomeDirectory("sauron v0.0.0\nHash abc\n")
	assert.Error(t, err)
}
