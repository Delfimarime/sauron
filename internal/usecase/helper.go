package usecase

// the selectable columns of a registry listing.
const (
	fieldName      = "name"
	fieldTransport = "transport"
	fieldURI       = "uri"
	fieldRef       = "ref"
	fieldTimeout   = "timeout"
)

// the sort directions a listing accepts.
const (
	orderAsc  = "asc"
	orderDesc = "desc"
)

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
