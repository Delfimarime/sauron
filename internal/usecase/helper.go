package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/registry/api"
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
