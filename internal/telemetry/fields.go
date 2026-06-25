package telemetry

// ECS-compatible custom field keys for Sauron's domain. A field that is not part
// of standard ECS is namespaced under the single custom top-level key "sauron",
// per ECS's custom-field convention.
const (
	FieldRegistryURI       = "sauron.registry.uri"
	FieldRegistryTransport = "sauron.registry.transport"
	FieldArtifactCount     = "sauron.artifact.count"
)
