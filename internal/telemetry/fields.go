package telemetry

// ECS-compatible custom field keys for Sauron's domain. A field that is not part
// of standard ECS is namespaced under the single custom top-level key "sauron",
// per ECS's custom-field convention.
const (
	FieldRegistrySource    = "sauron.registry.source"
	FieldRegistryTransport = "sauron.registry.transport"
	FieldArtifactCount     = "sauron.artifact.count"
	FieldArtifactName      = "sauron.artifact.name"
	FieldProviderName      = "sauron.provider.name"
	FieldProviderFrom      = "sauron.provider.from"
	FieldProviderTo        = "sauron.provider.to"
)
