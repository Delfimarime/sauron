//go:build unit

package gherkin

import (
	"context"
	"fmt"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// fakeRuntime is a runtime.Runtime stub for the gherkin controllers' unit tests. It
// returns canned Execute/ReadFile results, records the args it was called with, and
// hands out fakeSources keyed by capability+alias so resolver and fixture tests run
// without a host process or Docker.
type fakeRuntime struct {
	code int
	out  string
	err  error
	args []string

	files      map[string][]byte
	readErr    error
	folders    map[string]*fakeSource
	webservers map[string]*fakeSource
	gitErr     error
}

func (*fakeRuntime) IsReadOnly() bool            { return false }
func (*fakeRuntime) Start(context.Context) error { return nil }
func (*fakeRuntime) Stop(context.Context) error  { return nil }

func (f *fakeRuntime) Execute(_ context.Context, args ...string) (int, string, error) {
	f.args = args
	return f.code, f.out, f.err
}

func (f *fakeRuntime) CopyTo(context.Context, string, []byte) error { return nil }

func (f *fakeRuntime) ReadFile(_ context.Context, path string) ([]byte, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	if data, ok := f.files[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("fake: no file at %q", path)
}

func (f *fakeRuntime) Folder(alias string) runtime.Source    { return getOrAdd(&f.folders, alias) }
func (f *fakeRuntime) Webserver(alias string) runtime.Source { return getOrAdd(&f.webservers, alias) }

func (f *fakeRuntime) Git(string) runtime.Source {
	err := f.gitErr
	if err == nil {
		err = fmt.Errorf("fake: git source is deferred")
	}
	return runtime.NewErroringSource(err)
}

func getOrAdd(m *map[string]*fakeSource, alias string) *fakeSource {
	if *m == nil {
		*m = map[string]*fakeSource{}
	}
	src, ok := (*m)[alias]
	if !ok {
		src = &fakeSource{}
		(*m)[alias] = src
	}
	return src
}

// fakeSource records what was exposed and returns canned Path/URL/SSHKey/Revision
// values.
type fakeSource struct {
	path        string
	url         string
	sshKey      string
	revision    string
	pathErr     error
	urlErr      error
	sshKeyErr   error
	revisionErr error
	exposed     []runtime.Resource
}

func (s *fakeSource) Expose(resources ...runtime.Resource) {
	s.exposed = append(s.exposed, resources...)
}

func (s *fakeSource) Path(context.Context) (string, error) { return s.path, s.pathErr }

func (s *fakeSource) URL(context.Context) (string, error) { return s.url, s.urlErr }

func (s *fakeSource) SSHKey(context.Context) (string, error) { return s.sshKey, s.sshKeyErr }

func (s *fakeSource) Revision(context.Context) (string, error) { return s.revision, s.revisionErr }
