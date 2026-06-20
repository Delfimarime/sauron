package gherkin

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/delfimarime/sauron/pkg/sauron/types"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// skillResource builds one provider-content file in memory: a minimal Skill manifest
// under the ".skills/<name>/" prefix of the content set. The same resource is exposed
// through whichever source (folder, webserver) a fixture declares — one content set,
// three exposures.
func skillResource(name string) runtime.Resource {
	return runtime.Resource{
		Path:    ".skills/" + name + "/skill.yaml",
		Content: manifest(types.KindSkill, name),
	}
}

// manifest renders a minimal apiVersion/kind/metadata document for a content artifact.
func manifest(kind, name string) []byte {
	return []byte("apiVersion: " + types.APIVersion + "\n" +
		"kind: " + kind + "\n" +
		"metadata:\n" +
		"  name: " + name + "\n")
}

// exposeDirectory loads every file under root (a path relative to the e2e module
// root, e.g. "testdata/registries/acme") and exposes them on src, each served at its
// path relative to root. The bytes are read here and carried as inline Content, so
// the exposure is identical on the host runtime and inside the container — no
// dependence on the Docker daemon seeing the host path.
func exposeDirectory(src runtime.Source, root string) error {
	resources, err := collectResources(root)
	if err != nil {
		return err
	}
	src.Expose(resources...)
	return nil
}

// exposeFile loads a single file (path relative to the e2e module root) and exposes
// it on src at served.
func exposeFile(src runtime.Source, path, served string) error {
	resource, err := fileResource(path, served)
	if err != nil {
		return err
	}
	src.Expose(resource)
	return nil
}

// collectResources walks root and returns one content resource per file, each served
// at its slash-separated path relative to root. macOS ".DS_Store" noise is skipped. A
// missing or non-directory root is an error. Pure (a directory tree in, resources
// out) so it is unit-tested against a temp tree.
func collectResources(root string) ([]runtime.Resource, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("host the directory %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is a file, not a directory; use the file step", root)
	}

	var out []runtime.Resource
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == ".DS_Store" {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out = append(out, runtime.Resource{Path: filepath.ToSlash(rel), Content: content})
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %q: %w", root, walkErr)
	}
	return out, nil
}

// fileResource reads a single file and serves it at served, defaulting to the file's
// base name when served is empty.
func fileResource(path, served string) (runtime.Resource, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return runtime.Resource{}, fmt.Errorf("host the file %q: %w", path, err)
	}
	if served == "" {
		served = filepath.Base(path)
	}
	return runtime.Resource{Path: filepath.ToSlash(served), Content: content}, nil
}
