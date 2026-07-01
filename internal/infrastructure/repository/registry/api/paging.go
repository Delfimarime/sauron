package api

import (
	"sort"
	"strings"

	"github.com/delfimarime/sauron/pkg/sauron/source"
)

// Page applies a listing's search, sort and paging options to entries in that
// order: it filters by the case-insensitive Search term, orders by name
// (ascending unless Order is "desc"), then slices by Offset and Limit. When
// entries is non-nil, the returned slice is non-nil.
func Page(entries []source.File, options source.Options) []source.File {
	filtered := filter(entries, options)
	order(filtered, options)
	return slice(filtered, options)
}

// filter keeps only the entries whose name contains options.Search, comparing
// case-insensitively; a nil or empty Search matches every entry.
func filter(entries []source.File, options source.Options) []source.File {
	if options.Search == nil || *options.Search == "" {
		out := make([]source.File, len(entries))
		copy(out, entries)
		return out
	}

	term := strings.ToLower(*options.Search)
	matched := make([]source.File, 0, len(entries))
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name()), term) {
			matched = append(matched, entry)
		}
	}

	return matched
}

// order sorts entries in place by name; the direction is ascending unless
// options.Order is "desc".
func order(entries []source.File, options source.Options) {
	descending := options.Order != nil && *options.Order == "desc"
	sort.Slice(entries, func(i, j int) bool {
		if descending {
			return entries[i].Name() > entries[j].Name()
		}
		return entries[i].Name() < entries[j].Name()
	})
}

// slice applies the offset and limit from options to entries.
func slice(entries []source.File, options source.Options) []source.File {
	if options.Offset != nil {
		offset := int(*options.Offset)
		if offset >= len(entries) {
			return []source.File{}
		}
		entries = entries[offset:]
	}

	if options.Limit != nil {
		limit := int(*options.Limit)
		if limit >= 0 && limit < len(entries) {
			entries = entries[:limit]
		}
	}

	return entries
}
