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

// pagingLine renders the applied-paging report shared by every paginated
// listing view: an empty page reports zero results, a populated page the
// inclusive from–to window.
func pagingLine(page, limit, offset int64, count int) string {
	if count == 0 {
		return fmt.Sprintf("showing 0 results (page %d, limit %d)", page, limit)
	}

	from := offset + 1
	to := offset + int64(count)

	return fmt.Sprintf("showing %d–%d (page %d, limit %d)", from, to, page, limit)
}
