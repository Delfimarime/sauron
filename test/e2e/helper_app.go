package e2e

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
)

// App is the template context for feature files. Its fields mirror the build
// identity the gate injects via environment variables.
type App struct {
	Name        string
	Version     string
	Tag         string
	GitHash     string
	FullVersion string // Version-Tag
}

func determineTestdataDirectory(t *testing.T) string {
	app := App{
		Tag:     os.Getenv("SAURON_APP_TAG"),
		Name:    os.Getenv("SAURON_APP_NAME"),
		GitHash: os.Getenv("SAURON_GIT_HASH"),
		Version: os.Getenv("SAURON_APP_VERSION"),
	}
	app.FullVersion = app.Version + "-" + app.Tag
	features := t.TempDir()
	if err := renderFeatures("testdata", features, struct{ App App }{App: app}); err != nil {
		t.Fatalf("render feature templates: %s", err)
	}
	return features
}

// renderFeatures renders every *.feature file in srcDir as a text/template against
// data and writes the result into dstDir, so features can reference the build
// identity (e.g. {{.App.FullVersion}}) without hardcoding it. The srcDir files are
// left untouched; a reference to an unknown field is a hard error.
func renderFeatures(srcDir, dstDir string, data any) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read feature dir %q: %w", srcDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".feature") {
			continue
		}

		src := filepath.Join(srcDir, entry.Name())
		content, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read feature %q: %w", src, err)
		}

		tmpl, err := template.New(entry.Name()).Option("missingkey=error").Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse feature %q: %w", src, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("render feature %q: %w", src, err)
		}

		if err := os.WriteFile(filepath.Join(dstDir, entry.Name()), buf.Bytes(), 0o600); err != nil {
			return fmt.Errorf("write rendered feature %q: %w", entry.Name(), err)
		}
	}

	return nil
}

// resolveFeaturePaths returns the godog Paths to run: every rendered feature in
// dir when raw is empty, or only the rendered files named in raw (a
// comma-separated list of feature-file basenames) otherwise. An unknown name is
// a hard error — a typo should fail loudly, not silently run nothing extra.
func resolveFeaturePaths(dir, raw string) ([]string, error) {
	if raw == "" {
		return []string{dir}, nil
	}

	var paths []string
	for _, name := range strings.Split(raw, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("feature %q not found under %q: %w", name, dir, err)
		}
		paths = append(paths, path)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("-features was set but resolved to no valid feature files")
	}
	return paths, nil
}

// resolveConcurrency returns explicit if it's a positive override, or a
// CPU-aware default otherwise: min(5, max(1, runtime.NumCPU()/2)). Each
// scenario runs its own Docker Compose stack, so this leaves headroom for the
// Docker daemon rather than using every core for the test process itself.
func resolveConcurrency(explicit int) int {
	if explicit > 0 {
		return explicit
	}
	n := runtime.NumCPU() / 2
	if n < 1 {
		n = 1
	}
	if n > 5 {
		n = 5
	}
	return n
}
