package storage

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"gopkg.in/yaml.v3"
)

// errNoSchemaForKind reports that no schema is registered for a document kind.
var errNoSchemaForKind = errors.New("no schema registered for kind")

// schemaFiles holds the JSON schema for every document kind. The directory is
// staged by the build before compilation.
//
//go:embed schemas/*.json
var schemaFiles embed.FS

// validator validates a decoded document against the schema for its kind.
type validator struct {
	byKind map[string]*jsonschema.Resolved
}

// newValidator loads and resolves every embedded schema, keyed by the document
// kind it constrains (the schema's title).
func newValidator() (*validator, error) {
	entries, err := fs.ReadDir(schemaFiles, "schemas")
	if err != nil {
		return nil, fmt.Errorf("read embedded schemas: %w", err)
	}

	byKind := make(map[string]*jsonschema.Resolved, len(entries))
	for _, entry := range entries {
		kind, resolved, err := loadSchema(entry.Name())
		if err != nil {
			return nil, err
		}
		byKind[kind] = resolved
	}

	return &validator{byKind: byKind}, nil
}

// loadSchema reads, parses, and resolves a single embedded schema, returning the
// kind it constrains.
func loadSchema(name string) (string, *jsonschema.Resolved, error) {
	raw, err := schemaFiles.ReadFile("schemas/" + name)
	if err != nil {
		return "", nil, fmt.Errorf("read schema %q: %w", name, err)
	}

	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return "", nil, fmt.Errorf("parse schema %q: %w", name, err)
	}

	resolved, err := schema.Resolve(nil)
	if err != nil {
		return "", nil, fmt.Errorf("resolve schema %q: %w", name, err)
	}

	kind := schema.Title
	if kind == "" {
		kind = strings.TrimSuffix(name, ".schema.json")
	}

	return kind, resolved, nil
}

// validate decodes node to a JSON-shaped value and validates it against the
// schema registered for kind. An unknown kind is an error.
func (v *validator) validate(kind string, node *yaml.Node) error {
	resolved, ok := v.byKind[kind]
	if !ok {
		return fmt.Errorf("%w %q", errNoSchemaForKind, kind)
	}

	var instance any
	if err := node.Decode(&instance); err != nil {
		return fmt.Errorf("decode %s document: %w", kind, err)
	}

	if err := resolved.Validate(instance); err != nil {
		return fmt.Errorf("validate %s document: %w", kind, err)
	}

	return nil
}
