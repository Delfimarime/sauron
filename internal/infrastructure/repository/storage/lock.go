package storage

import (
	"fmt"
	"sync"

	"github.com/spf13/afero"
)

// lockFile is the guard file created under the home directory while a write is
// in flight.
const lockFile = ".lock"

// guard serializes writers over the home filesystem. An in-process mutex orders
// goroutines; an on-disk lock file signals the critical section to other
// processes sharing the same home.
type guard struct {
	fs afero.Fs
	mu sync.Mutex
}

// newGuard builds a guard over fs.
func newGuard(fs afero.Fs) *guard {
	return &guard{fs: fs}
}

// withLock runs fn while holding the write lock, releasing it afterwards even if
// fn fails.
func (g *guard) withLock(fn func() error) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if err := g.acquire(); err != nil {
		return err
	}
	defer g.release()

	return fn()
}

// acquire creates the on-disk lock file.
func (g *guard) acquire() error {
	f, err := g.fs.OpenFile(lockFile, openExclusive, lockPerm)
	if err != nil {
		return fmt.Errorf("acquire write lock: %w", err)
	}
	return f.Close()
}

// release removes the on-disk lock file.
func (g *guard) release() {
	_ = g.fs.Remove(lockFile)
}
