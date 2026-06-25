package cmd

import "fmt"

// CataloguePagingLine renders the applied-paging report: an empty page reports
// zero results, a populated page the inclusive from–to window.
func CataloguePagingLine(page, limit int64, count int) string {
	if count == 0 {
		return fmt.Sprintf("showing 0 results (page %d, limit %d)", page, limit)
	}

	offset := (page - 1) * limit
	from := offset + 1
	to := offset + int64(count)

	return fmt.Sprintf("showing %d–%d (page %d, limit %d)", from, to, page, limit)
}
