package cmd

import (
	"fmt"
	"io"

	"github.com/delfimarime/sauron/internal/usecase"
)

// errWriter is a write helper that accumulates the first error and silently
// no-ops every subsequent write, so renderers can issue multiple printf calls
// without per-call error checks. Call toIOError once at the end.
type errWriter struct {
	w   io.Writer
	err error
}

// newErrWriter wraps w in a sticky-error writer.
func newErrWriter(w io.Writer) *errWriter {
	return &errWriter{w: w}
}

// printf formats and writes like fmt.Fprintf; once an error has been recorded
// all subsequent calls are silent no-ops.
func (ew *errWriter) printf(format string, args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, args...)
}

// record notes an external error (e.g. from a sub-renderer) without writing;
// once an error is recorded all subsequent printf calls are silent no-ops.
func (ew *errWriter) record(err error) {
	if ew.err == nil {
		ew.err = err
	}
}

// toIOError returns nil when no write has failed, or a classified io error
// with context prepended.
func (ew *errWriter) toIOError(context string) error {
	if ew.err == nil {
		return nil
	}
	return usecase.NewIOError(fmt.Sprintf("%s: %v", context, ew.err))
}
