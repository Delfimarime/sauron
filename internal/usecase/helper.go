package usecase

// defaultSortOrder applies the listing defaults: an empty sort becomes fieldName
// and an empty order becomes orderAsc.
func defaultSortOrder(sort, order string) (string, string) {
	if sort == "" {
		sort = fieldName
	}
	if order == "" {
		order = orderAsc
	}

	return sort, order
}

// isValidOrder reports whether order is one of the accepted sort directions.
func isValidOrder(order string) bool {
	return order == orderAsc || order == orderDesc
}

// fieldName is the catalogue's only sortable key.
const fieldName = "name"

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
