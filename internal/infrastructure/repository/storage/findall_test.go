package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// betaRegistryYAML is a second schema-valid Registry document.
const betaRegistryYAML = `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: beta
spec:
  transport: http
  uri: https://example.com/beta
`

// TestStoreFindAllReturnsEveryDocument returns every document of the kind, in order.
func TestStoreFindAllReturnsEveryDocument(t *testing.T) {
	// Arrange: two documents in the stream.
	store, _ := newTestStore(t)
	require.NoError(t, store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML)))
	require.NoError(t, store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, betaRegistryYAML)))

	// Act.
	docs, err := store.FindAll(context.Background(), types.KindRegistry)

	// Assert.
	require.NoError(t, err)
	require.Len(t, docs, 2)
	assert.Equal(t, acmeName, nameOf(docs[0]))
	assert.Equal(t, "beta", nameOf(docs[1]))
}

// TestStoreFindAllMissingFileIsEmpty returns an empty slice, not an error, when the
// backing file is absent.
func TestStoreFindAllMissingFileIsEmpty(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	docs, err := store.FindAll(context.Background(), types.KindRegistry)

	// Assert.
	require.NoError(t, err)
	assert.Empty(t, docs)
}

// TestStoreFindAllUnknownKind reports an error for a kind with no backing file.
func TestStoreFindAllUnknownKind(t *testing.T) {
	// Arrange.
	store, _ := newTestStore(t)

	// Act.
	_, err := store.FindAll(context.Background(), "Nonexistent")

	// Assert.
	require.ErrorIs(t, err, errUnknownKind)
}

// TestStoreFindAllMalformedStream surfaces a YAML parse error from the file.
func TestStoreFindAllMalformedStream(t *testing.T) {
	// Arrange: invalid YAML in the registries file.
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte("\tnot: [valid"), filePerm))

	// Act.
	_, err := store.FindAll(context.Background(), types.KindRegistry)

	// Assert.
	require.Error(t, err)
}

// TestStoreFindAllValidatesEveryDocument fails the whole read when any single
// document violates its schema (validation is all-or-nothing).
func TestStoreFindAllValidatesEveryDocument(t *testing.T) {
	// Arrange: a valid document followed by one with an unknown spec field
	// (additionalProperties: false).
	invalid := `apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: beta
spec:
  transport: git
  uri: https://example.com/beta.git
  bogus: nope
`
	store, fs := newTestStore(t)
	require.NoError(t, afero.WriteFile(fs, registriesFile, []byte(validRegistryYAML+documentSeparator+invalid), filePerm))

	// Act.
	_, err := store.FindAll(context.Background(), types.KindRegistry)

	// Assert.
	require.Error(t, err)
}

// errRenameFail is returned by renameFailFs to drive the writeAtomic commit branch.
var errRenameFail = errors.New("rename failed")

// renameFailFs is an afero.Fs whose Rename always fails, so the temp file written
// by writeAtomic cannot be committed into place.
type renameFailFs struct {
	afero.Fs
}

// Rename always fails, exercising the writeAtomic cleanup path.
func (renameFailFs) Rename(_, _ string) error {
	return errRenameFail
}

// TestStoreWriteAtomicRenameFailureCleansTemp surfaces a commit error and removes
// the temp file when the rename into place fails.
func TestStoreWriteAtomicRenameFailureCleansTemp(t *testing.T) {
	// Arrange: a filesystem that accepts the temp write but rejects the rename.
	fs := renameFailFs{afero.NewMemMapFs()}
	store, err := NewStore(fs)
	require.NoError(t, err)

	// Act.
	err = store.Append(context.Background(), types.KindRegistry, nodeFromYAML(t, validRegistryYAML))

	// Assert: the commit fails and the temp artifact is cleaned up.
	require.Error(t, err)
	exists, existsErr := afero.Exists(fs, registriesFile+".tmp")
	require.NoError(t, existsErr)
	assert.False(t, exists)
}
