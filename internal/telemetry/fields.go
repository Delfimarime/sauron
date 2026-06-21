package telemetry

// ECS-compatible custom field keys for Sauron's domain. A field that is not part
// of standard ECS is namespaced under the single custom top-level key "sauron",
// per ECS's custom-field convention.
const (
	FieldRegistryName      = "sauron.registry.name"
	FieldRegistryTransport = "sauron.registry.transport"
	FieldRegistryCount     = "sauron.registry.count"
	FieldArtifactCount     = "sauron.artifact.count"
)
