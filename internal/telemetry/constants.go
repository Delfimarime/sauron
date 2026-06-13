package telemetry

// Shared ECS field keys referenced by structured logs; a seed set features extend.
const (
	FieldEventAction    = "event.action"
	FieldEventOutcome   = "event.outcome"
	FieldErrorMessage   = "error.message"
	FieldFilePath       = "file.path"
	FieldServiceName    = "service.name"
	FieldServiceVersion = "service.version"
)
