// Package storage owns Sauron's persisted state under the configured home.
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

const (
	// openExclusive opens a file for writing, failing if it already exists.
	openExclusive = os.O_CREATE | os.O_EXCL | os.O_WRONLY
	// filePerm is the mode for persisted document files.
	filePerm os.FileMode = 0o644
	// lockPerm is the mode for the write lock file.
	lockPerm os.FileMode = 0o600
)

// errUnknownKind reports a kind with no backing file.
var errUnknownKind = errors.New("unknown document kind")

// Store reads and writes Sauron's persisted state over the injected filesystem.
// It is kind-agnostic: each document kind maps to a file holding a multi-document
// YAML stream.
type Store struct {
	fs        afero.Fs
	guard     *guard
	validator *validator
	files     map[string]string
}

// NewStore builds a Store over fs.
func NewStore(fs afero.Fs) (*Store, error) {
	v, err := newValidator()
	if err != nil {
		return nil, err
	}

	return &Store{
		fs:        fs,
		guard:     newGuard(fs),
		validator: v,
		files:     map[string]string{types.KindRegistry: "registries.yaml"},
	}, nil
}

// fileFor resolves the backing file for kind.
func (s *Store) fileFor(kind string) (string, error) {
	name, ok := s.files[kind]
	if !ok {
		return "", fmt.Errorf("%w: %q", errUnknownKind, kind)
	}
	return name, nil
}

// FindOne returns the document of the given kind whose metadata.name matches
// name, validated against its schema. It returns nil when the file or a matching
// document is absent.
func (s *Store) FindOne(_ context.Context, kind, name string) (*yaml.Node, error) {
	file, err := s.fileFor(kind)
	if err != nil {
		return nil, err
	}

	docs, err := s.readDocuments(file)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		if nameOf(doc) != name {
			continue
		}
		if err := s.validator.validate(kind, doc); err != nil {
			return nil, err
		}
		return doc, nil
	}

	return nil, nil
}

// FindAll returns every document of the given kind, each validated against its
// schema. Validation is all-or-nothing: a single schema-invalid document fails
// the whole read. A missing file yields an empty slice, not an error.
func (s *Store) FindAll(_ context.Context, kind string) ([]*yaml.Node, error) {
	file, err := s.fileFor(kind)
	if err != nil {
		return nil, err
	}

	docs, err := s.readDocuments(file)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		if err := s.validator.validate(kind, doc); err != nil {
			return nil, err
		}
	}

	return docs, nil
}

// Append atomically adds doc to the kind's file under the write lock. The
// document is not re-validated on write.
func (s *Store) Append(_ context.Context, kind string, doc *yaml.Node) error {
	file, err := s.fileFor(kind)
	if err != nil {
		return err
	}

	return s.guard.withLock(func() error {
		return s.appendLocked(file, doc)
	})
}

// appendLocked performs the read-modify-write append while the lock is held.
func (s *Store) appendLocked(file string, doc *yaml.Node) error {
	existing, err := s.readRaw(file)
	if err != nil {
		return err
	}

	encoded, err := encodeDocument(doc)
	if err != nil {
		return err
	}

	return s.writeAtomic(file, append(existing, encoded...))
}

// readDocuments parses the multi-document YAML stream in file. A missing file
// yields no documents.
func (s *Store) readDocuments(file string) ([]*yaml.Node, error) {
	raw, err := s.readRaw(file)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}

	var docs []*yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	for {
		var node yaml.Node
		err := decoder.Decode(&node)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", file, err)
		}
		docs = append(docs, &node)
	}

	return docs, nil
}

// readRaw returns the raw bytes of file, or nil when it does not exist.
func (s *Store) readRaw(file string) ([]byte, error) {
	raw, err := afero.ReadFile(s.fs, file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	return raw, nil
}

// writeAtomic writes data to a temporary file and renames it into place.
func (s *Store) writeAtomic(file string, data []byte) error {
	tmp := file + ".tmp"
	if err := afero.WriteFile(s.fs, tmp, data, filePerm); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := s.fs.Rename(tmp, file); err != nil {
		_ = s.fs.Remove(tmp)
		return fmt.Errorf("commit %s: %w", file, err)
	}
	return nil
}

// nameOf returns the metadata.name of a document node, or "" when absent.
func nameOf(doc *yaml.Node) string {
	var envelope struct {
		Metadata struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
	}
	if err := doc.Decode(&envelope); err != nil {
		return ""
	}
	return envelope.Metadata.Name
}

// documentSeparator delimits documents in a multi-document YAML stream.
const documentSeparator = "---\n"

// encodeDocument renders doc as a single YAML document body, prefixed with a
// stream separator so it appends cleanly after existing documents.
func encodeDocument(doc *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(documentBody(doc)); err != nil {
		_ = encoder.Close()
		return nil, fmt.Errorf("encode document: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("encode document: %w", err)
	}
	return append([]byte(documentSeparator), buf.Bytes()...), nil
}

// documentBody unwraps a DocumentNode to its content so encoding does not nest a
// second document wrapper.
func documentBody(doc *yaml.Node) *yaml.Node {
	if doc.Kind == yaml.DocumentNode && len(doc.Content) == 1 {
		return doc.Content[0]
	}
	return doc
}
