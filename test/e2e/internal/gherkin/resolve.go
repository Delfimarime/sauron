package gherkin

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// defaultAlias is assumed when a #{} reference omits the alias segment.
const defaultAlias = "default"

// valueOf resolves a step argument to T. A "#{.capability[.alias].attr}" string is
// routed to the runtime's typed source accessor (forcing the lazy Start); any other
// value is coerced directly. This is the ONLY resolver — the runtime owns the
// provisioned addresses, this helper owns the parsing and typing.
func valueOf[T any](ctx context.Context, rt runtime.Runtime, raw any) (T, error) {
	var zero T
	s, ok := raw.(string)
	if !ok {
		if v, ok := raw.(T); ok {
			return v, nil
		}
		return zero, fmt.Errorf("valueOf: cannot use %T as %T", raw, zero)
	}
	if isReference(s) {
		resolved, err := resolveReference(ctx, rt, s)
		if err != nil {
			return zero, err
		}
		return coerce[T](resolved)
	}
	return coerce[T](s)
}

// isReference reports whether s is a #{…} dynamic reference.
func isReference(s string) bool {
	return strings.HasPrefix(s, "#{") && strings.HasSuffix(s, "}")
}

// reference is a parsed #{.capability[.alias].attr} expression.
type reference struct {
	capability string
	alias      string
	attr       string
}

// parseReference splits a #{…} expression. The leading dot is mandatory; a missing
// alias defaults to "default".
func parseReference(s string) (reference, error) {
	inner := strings.TrimSuffix(strings.TrimPrefix(s, "#{"), "}")
	if !strings.HasPrefix(inner, ".") {
		return reference{}, fmt.Errorf("reference %q must start with a dot: #{.capability[.alias].attr}", s)
	}
	parts := strings.Split(strings.TrimPrefix(inner, "."), ".")
	switch len(parts) {
	case 2:
		return reference{capability: parts[0], alias: defaultAlias, attr: parts[1]}, nil
	case 3:
		return reference{capability: parts[0], alias: parts[1], attr: parts[2]}, nil
	default:
		return reference{}, fmt.Errorf("malformed reference %q: want #{.capability[.alias].attr}", s)
	}
}

// resolveReference routes a parsed reference to the runtime accessor that owns its
// address. The attribute must match the capability's headline attribute.
func resolveReference(ctx context.Context, rt runtime.Runtime, s string) (string, error) {
	ref, err := parseReference(s)
	if err != nil {
		return "", err
	}
	switch ref.capability {
	case "folder":
		if ref.attr != "path" {
			return "", fmt.Errorf("folder reference %q: only .path is supported", s)
		}
		return rt.Folder(ref.alias).Path(ctx)
	case "webserver":
		if ref.attr != "url" {
			return "", fmt.Errorf("webserver reference %q: only .url is supported", s)
		}
		return rt.Webserver(ref.alias).URL(ctx)
	case "git":
		switch ref.attr {
		case "url":
			return rt.Git(ref.alias).URL(ctx)
		case "sshKey":
			return rt.Git(ref.alias).SSHKey(ctx)
		case "revision":
			return rt.Git(ref.alias).Revision(ctx)
		default:
			return "", fmt.Errorf("git reference %q: only .url, .sshKey, and .revision are supported", s)
		}
	default:
		return "", fmt.Errorf("unknown capability %q in reference %q", ref.capability, s)
	}
}

// coerce converts a resolved string to the requested target type. Only the types the
// step catalog needs (string, int) are supported.
func coerce[T any](s string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case string:
		return any(s).(T), nil
	case int:
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return zero, fmt.Errorf("coerce %q to int: %w", s, err)
		}
		return any(n).(T), nil
	default:
		return zero, fmt.Errorf("valueOf: unsupported target type %T", zero)
	}
}
