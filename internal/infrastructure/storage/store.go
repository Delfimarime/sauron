// Package storage owns Sauron's persisted state under the configured home.
package storage

import (
	"github.com/spf13/afero"
)

// Store reads and writes Sauron's persisted state over the injected filesystem.
type Store struct {
	fs afero.Fs
}

// NewStore builds a Store over fs.
func NewStore(fs afero.Fs) *Store {
	return &Store{fs: fs}
}
