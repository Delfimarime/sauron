package marketplace

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError is a non-2xx response from the registry, carrying the problem detail
// the server returned (RFC 9457 / application/problem+json) when present.
type APIError struct {
	Status int
	Type   string
	Title  string
	Detail string
}

// Error renders the API error for humans.
func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("registry responded %d: %s", e.Status, e.Detail)
	}
	if e.Title != "" {
		return fmt.Sprintf("registry responded %d: %s", e.Status, e.Title)
	}
	return fmt.Sprintf("registry responded %d", e.Status)
}

// ErrInvalidConfig marks a client that cannot be built from the supplied options.
var ErrInvalidConfig = errors.New("invalid marketplace configuration")

// ErrTransport marks a failure to reach the registry, before any HTTP status was
// observed.
var ErrTransport = errors.New("marketplace transport failed")

// IsNotFound reports whether err is an APIError with a 404 status.
func IsNotFound(err error) bool {
	return hasStatus(err, http.StatusNotFound)
}

// IsUnauthorized reports whether err is an APIError with a 401 status.
func IsUnauthorized(err error) bool {
	return hasStatus(err, http.StatusUnauthorized)
}

// IsForbidden reports whether err is an APIError with a 403 status.
func IsForbidden(err error) bool {
	return hasStatus(err, http.StatusForbidden)
}

// IsBadRequest reports whether err is an APIError with a 400 status.
func IsBadRequest(err error) bool {
	return hasStatus(err, http.StatusBadRequest)
}

// hasStatus reports whether err unwraps to an APIError carrying the given status.
func hasStatus(err error, status int) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == status
}
