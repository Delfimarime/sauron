package api

import "errors"

// ErrUsage marks a misconfiguration the caller can correct, such as supplying an
// option a transport does not accept.
var ErrUsage = errors.New("invalid registry configuration")

// ErrRuntime marks a failure encountered while opening or reading a source, such
// as an unreachable host or an unreadable location.
var ErrRuntime = errors.New("registry access failed")
