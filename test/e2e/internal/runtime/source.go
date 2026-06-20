package runtime

import "context"

// ErroringSource is a Source for a capability a backend does not offer. Expose is a
// silent no-op (Given steps accumulate), but any attribute access (Path/URL) — the
// point where a scenario actually needs the address — returns err. This is how a
// capability gap surfaces honestly, without a type-assert.
type ErroringSource struct{ Err error }

// NewErroringSource returns a Source whose Path/URL always fail with err.
func NewErroringSource(err error) Source { return ErroringSource{Err: err} }

func (ErroringSource) Expose(...Resource) {}

func (s ErroringSource) Path(context.Context) (string, error) { return "", s.Err }

func (s ErroringSource) URL(context.Context) (string, error) { return "", s.Err }

func (s ErroringSource) SSHKey(context.Context) (string, error) { return "", s.Err }
