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
	// settingsFile holds the singleton Registry and Provider documents.
	settingsFile = "settings.yaml"
	// trackFile holds the installed Skill and Agent documents.
	trackFile = "track.yaml"
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
		files: map[string]string{
			types.KindRegistry: settingsFile,
			types.KindProvider: settingsFile,
			types.KindSkill:    trackFile,
			types.KindAgent:    trackFile,
		},
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

	matches := make([]*yaml.Node, 0, len(docs))
	for _, doc := range docs {
		if kindOf(doc) != kind {
			continue
		}
		if err := s.validator.validate(kind, doc); err != nil {
			return nil, err
		}
		matches = append(matches, doc)
	}

	return matches, nil
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

// Remove atomically drops the document of the given kind whose metadata.name
// matches name from the kind's file, under the write lock. Removing an absent
// document (or from a missing file) is a no-op success.
func (s *Store) Remove(_ context.Context, kind, name string) error {
	file, err := s.fileFor(kind)
	if err != nil {
		return err
	}

	return s.guard.withLock(func() error {
		return s.removeLocked(file, name)
	})
}

// removeLocked performs the read-filter-write removal while the lock is held. When
// no document matches, the file is left untouched.
func (s *Store) removeLocked(file, name string) error {
	docs, err := s.readDocuments(file)
	if err != nil {
		return err
	}

	kept := make([]*yaml.Node, 0, len(docs))
	for _, doc := range docs {
		if nameOf(doc) == name {
			continue
		}
		kept = append(kept, doc)
	}
	if len(kept) == len(docs) {
		return nil
	}

	stream, err := encodeStream(kept)
	if err != nil {
		return err
	}

	return s.writeAtomic(file, stream)
}

// First returns the first document of kind in its file, validated against the
// kind's schema, or nil when none is present. It filters by the document's own
// kind, so a file that holds several kinds (settings.yaml) yields only matches.
func (s *Store) First(_ context.Context, kind string) (*yaml.Node, error) {
	file, err := s.fileFor(kind)
	if err != nil {
		return nil, err
	}

	docs, err := s.readDocuments(file)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		if kindOf(doc) != kind {
			continue
		}
		if err := s.validator.validate(kind, doc); err != nil {
			return nil, err
		}
		return doc, nil
	}

	return nil, nil
}

// Replace atomically rewrites the kind's file so it holds exactly doc for that
// kind, preserving every document of another kind in the same file, under the
// write lock. The document is not re-validated on write.
func (s *Store) Replace(_ context.Context, kind string, doc *yaml.Node) error {
	file, err := s.fileFor(kind)
	if err != nil {
		return err
	}

	return s.guard.withLock(func() error {
		return s.replaceLocked(file, kind, doc)
	})
}

// replaceLocked drops every document of kind and appends the new one while the
// lock is held, keeping documents of other kinds untouched.
func (s *Store) replaceLocked(file, kind string, doc *yaml.Node) error {
	docs, err := s.readDocuments(file)
	if err != nil {
		return err
	}

	kept := make([]*yaml.Node, 0, len(docs)+1)
	for _, d := range docs {
		if kindOf(d) == kind {
			continue
		}
		kept = append(kept, d)
	}
	kept = append(kept, doc)

	stream, err := encodeStream(kept)
	if err != nil {
		return err
	}

	return s.writeAtomic(file, stream)
}

// Upsert atomically replaces the document of kind whose metadata.name matches
// name with doc, or appends doc when none matches, preserving every other
// document in the file, under the write lock. The document is not re-validated
// on write.
func (s *Store) Upsert(_ context.Context, kind, name string, doc *yaml.Node) error {
	file, err := s.fileFor(kind)
	if err != nil {
		return err
	}

	return s.guard.withLock(func() error {
		return s.upsertLocked(file, kind, name, doc)
	})
}

// upsertLocked replaces the matching (kind, name) document with doc, or appends
// it when absent, while the lock is held, keeping every other document untouched.
func (s *Store) upsertLocked(file, kind, name string, doc *yaml.Node) error {
	docs, err := s.readDocuments(file)
	if err != nil {
		return err
	}

	replaced := false
	kept := make([]*yaml.Node, 0, len(docs)+1)
	for _, d := range docs {
		if kindOf(d) == kind && nameOf(d) == name {
			kept = append(kept, doc)
			replaced = true
			continue
		}
		kept = append(kept, d)
	}
	if !replaced {
		kept = append(kept, doc)
	}

	stream, err := encodeStream(kept)
	if err != nil {
		return err
	}

	return s.writeAtomic(file, stream)
}

// Purge atomically removes every document of kind from its file, preserving
// documents of other kinds, under the write lock. Purging when none is present
// is a no-op success that leaves the file untouched.
func (s *Store) Purge(_ context.Context, kind string) error {
	file, err := s.fileFor(kind)
	if err != nil {
		return err
	}

	return s.guard.withLock(func() error {
		return s.purgeLocked(file, kind)
	})
}

// purgeLocked drops every document of kind while the lock is held; when none
// matches, the file is left byte-for-byte untouched.
func (s *Store) purgeLocked(file, kind string) error {
	docs, err := s.readDocuments(file)
	if err != nil {
		return err
	}

	kept := make([]*yaml.Node, 0, len(docs))
	for _, d := range docs {
		if kindOf(d) == kind {
			continue
		}
		kept = append(kept, d)
	}
	if len(kept) == len(docs) {
		return nil
	}

	stream, err := encodeStream(kept)
	if err != nil {
		return err
	}

	return s.writeAtomic(file, stream)
}

// encodeStream renders documents as a multi-document YAML stream, each prefixed
// with a stream separator — the same on-disk shape Append produces. An empty slice
// yields no bytes, so removing the last document leaves an empty file.
func encodeStream(docs []*yaml.Node) ([]byte, error) {
	var stream []byte
	for _, doc := range docs {
		encoded, err := encodeDocument(doc)
		if err != nil {
			return nil, err
		}
		stream = append(stream, encoded...)
	}

	return stream, nil
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

// kindOf returns the kind of a document node, or "" when absent. It lets the
// store operate on one kind within a file shared by several kinds.
func kindOf(doc *yaml.Node) string {
	var envelope struct {
		Kind string `yaml:"kind"`
	}
	if err := doc.Decode(&envelope); err != nil {
		return ""
	}
	return envelope.Kind
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
