// Package telemetry defines the ECS field keys shared across Sauron's tiers, so
// the logging vocabulary is declared once on the public surface. Public packages
// (e.g. pkg/http) and internal/telemetry both reference these constants rather
// than scattering string literals; internal-only keys, when any, stay in
// internal/telemetry.
package telemetry

// Common ECS field keys.
const (
	FieldEventAction    = "event.action"
	FieldEventOutcome   = "event.outcome"
	FieldErrorMessage   = "error.message"
	FieldFilePath       = "file.path"
	FieldServiceName    = "service.name"
	FieldServiceVersion = "service.version"
)

// HTTP log structure keys emitted by pkg/http.
const (
	KeyHTTP       = "http"
	KeyRequest    = "request"
	KeyResponse   = "response"
	KeyMethod     = "method"
	KeyURL        = "url"
	KeyHost       = "host"
	KeyMimeType   = "mime_type"
	KeyVersion    = "version"
	KeyStatusCode = "status_code"
)

// HeaderContentType is the HTTP Content-Type header name.
const HeaderContentType = "Content-Type"
