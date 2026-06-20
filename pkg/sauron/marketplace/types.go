package marketplace

// ArtifactSummary is the condensed view of an artifact returned in a listing. Its
// version and size are nullable: the registry omits them when it declares none.
type ArtifactSummary struct {
	Name    string  `json:"name"`
	Version *string `json:"version"`
	Size    *int64  `json:"size"`
}

// ArtifactList is one page of artifact summaries.
type ArtifactList struct {
	Items []ArtifactSummary `json:"items"`
}

// ListOptions configures a listing request.
type ListOptions struct {
	Search *string
	Sort   *string
	Limit  *int64
	Offset *int64
}

// ListOption mutates ListOptions.
type ListOption func(*ListOptions)

// WithSearch filters the listing by a case-insensitive substring of the name.
func WithSearch(search string) ListOption {
	return func(o *ListOptions) {
		o.Search = &search
	}
}

// WithSort orders the listing by the given directive (for example "+name").
func WithSort(sort string) ListOption {
	return func(o *ListOptions) {
		o.Sort = &sort
	}
}

// WithLimit caps the number of items returned in one page.
func WithLimit(limit int64) ListOption {
	return func(o *ListOptions) {
		o.Limit = &limit
	}
}

// WithOffset skips the given number of items before collecting the page.
func WithOffset(offset int64) ListOption {
	return func(o *ListOptions) {
		o.Offset = &offset
	}
}
