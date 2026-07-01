package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/extension"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// dirPerm is the mode for created provider directories.
const dirPerm os.FileMode = 0o755

// filePerm is the mode for files written into the provider filesystem.
const filePerm os.FileMode = 0o644

// listPageSize is the window listAll requests per page; sources default to paging
// (the http transport caps at 50), so a fixed window walks every entry.
const listPageSize int64 = 50

// providerDirs maps a provider name to the home-relative directory its artifacts
// live under.
var providerDirs = map[string]string{
	types.ProviderClaude:   ".claude",
	types.ProviderZencoder: ".zencoder",
}

// artifactKindDirs maps a document Kind to the registry root directory that
// holds artifacts of that type. Mirrors the provider home layout used by
// migrate's providerDirs.
var artifactKindDirs = map[string]string{
	types.KindSkill: "skills",
	types.KindAgent: "agents",
}

// installPath returns the path relative to the provider home where an artifact
// of the given kind and name is installed: "<kindDir>/sauron-<name>".
func installPath(kind, name string) string {
	return fmt.Sprintf("%s/sauron-%s", artifactKindDirs[kind], name)
}

// fetchURI builds the source URI for an artifact of the given kind and name.
func fetchURI(kind, name string) string {
	return artifactKindDirs[kind] + "/" + name
}

// trackKey is the reconciliation key for a recorded or desired artifact.
func trackKey(kind, name string) string {
	return kind + "/" + name
}

// listAll pages the source listing under root until it is exhausted, returning
// every entry. Sources page by default (the http transport caps a page at 50), so
// a single List call can miss entries; listAll walks fixed windows until a short
// page signals the end. Reused by install and sync to resolve desired versions
// without reading any artifact content.
func listAll(ctx context.Context, fs source.FileSystem, root string) ([]source.File, error) {
	var all []source.File
	for offset := int64(0); ; offset += listPageSize {
		page, err := fs.List(ctx, root,
			source.WithOffset(offset), source.WithLimit(listPageSize),
		)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if int64(len(page)) < listPageSize {
			return all, nil
		}
	}
}

// catalogueName trims a manifest's .yaml or .yml extension to its catalogue name.
func catalogueName(filename string) string {
	for _, ext := range []string{".yaml", ".yml"} {
		if strings.HasSuffix(filename, ext) {
			return strings.TrimSuffix(filename, ext)
		}
	}

	return filename
}

// ioErr wraps a storage failure as a classified io *Error, prefixing the reason
// with verb: a single shape for the "read/persist/remove X: <err>" store wraps.
func ioErr(verb string, err error) *Error {
	return NewIOError(fmt.Sprintf("%s: %v", verb, err))
}

// requireRegistry resolves the single configured registry, classifying a read
// failure as io and a missing registry as not-found. It is the shared
// "no registry is configured" guard install, list-catalogue, and
// describe-registry compose; it is the single source for that literal.
func requireRegistry(ctx context.Context, registries storage.RegistriesStore) (*types.Registry, error) {
	registry, err := registries.Get(ctx)
	if err != nil {
		return nil, ioErr("read registry", err)
	}
	if registry == nil {
		return nil, NewNotFoundError("no registry is configured")
	}

	return registry, nil
}

// identityRef returns value unchanged. It is the credential-resolution strategy
// the validate path uses to keep ${env:VAR} references verbatim, in contrast to
// the open path, which resolves them to their environment values.
func identityRef(value string) (string, error) {
	return value, nil
}

// connectOptions builds the extension option set from a registry spec, resolving
// each credential through resolve — identityRef to keep references verbatim for
// validation, the open use case's env-resolving strategy to connect live. It is
// the shared option-builder the set-registry validate and open-registry connect
// paths compose; their only difference is the resolve strategy.
func connectOptions(spec types.RegistrySpec, resolve func(string) (string, error)) ([]extension.Option, error) {
	opts := []extension.Option{extension.WithURI(spec.Source)}

	if spec.Transport == types.TransportGit && spec.Revision != "" {
		opts = append(opts, extension.WithRef(spec.Revision))
	}
	if timeout, err := parseTimeout(spec.Timeout); err != nil {
		return nil, err
	} else if timeout > 0 {
		opts = append(opts, extension.WithTimeout(timeout))
	}
	if spec.SSHKey != "" {
		opts = append(opts, extension.WithSSHKey(spec.SSHKey))
	}

	credentialsOpt, err := credentialsOption(spec.Credentials, resolve)
	if err != nil {
		return nil, err
	}
	if credentialsOpt != nil {
		opts = append(opts, credentialsOpt)
	}

	return append(opts, tlsOptions(spec.TLS)...), nil
}

// parseTimeout parses the spec's Go duration string; an empty value yields no
// bound. An unparsable value is a usage error.
func parseTimeout(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, NewUsageError(fmt.Sprintf("invalid timeout %q: %v", value, err))
	}

	return timeout, nil
}

// credentialsOption builds the basic-auth option, resolving each credential
// through resolve; it returns nil when no credentials were supplied.
func credentialsOption(
	credentials *types.Credentials, resolve func(string) (string, error),
) (extension.Option, error) {
	if credentials == nil || (credentials.Username == "" && credentials.Password == "") {
		return nil, nil
	}

	username, err := resolve(credentials.Username)
	if err != nil {
		return nil, err
	}
	password, err := resolve(credentials.Password)
	if err != nil {
		return nil, err
	}

	return extension.WithBasicAuth(username, password), nil
}

// tlsOptions builds the transport-security options from the spec block; a nil
// block yields none. Shared by the set-registry validate and open-registry
// connect paths.
func tlsOptions(tls *types.TLS) []extension.Option {
	if tls == nil {
		return nil
	}

	var opts []extension.Option
	if tls.SkipVerify {
		opts = append(opts, extension.WithSkipTLSVerify(true))
	}
	if tls.CACert != "" {
		opts = append(opts, extension.WithCACert(tls.CACert))
	}
	if tls.ClientCert != "" || tls.ClientKey != "" {
		opts = append(opts, extension.WithClientCert(tls.ClientCert, tls.ClientKey))
	}

	return opts
}

// classifyAdapterErr maps an adapter failure to a classified use-case error: a
// usage class is preserved, anything else becomes unreachable.
func classifyAdapterErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, api.ErrUsage) {
		return NewUsageError(err.Error())
	}

	return NewUnreachableError(err.Error())
}

// predicate reports whether an item should be kept.
type predicate[T any] func(T) bool

// filterBy keeps the items the predicate accepts.
func filterBy[T any](items []T, keep predicate[T]) []T {
	kept := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			kept = append(kept, item)
		}
	}

	return kept
}
