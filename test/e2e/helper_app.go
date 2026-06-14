package e2e

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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
