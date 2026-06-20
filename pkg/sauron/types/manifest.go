// Package types holds the public Go form of sauron's on-disk configuration
// documents. Each type is a faithful, behaviour-free DTO mirroring a persisted
// document schema; field tags track the schema property names exactly so the
// documents round-trip through YAML unchanged.
package types

// APIVersion is the only apiVersion sauron's v1 documents carry.
const APIVersion = "sauron.raitonbl.com/v1"

// Document kinds, one per persisted document schema.
const (
	KindRegistry = "Registry"
	KindSkill    = "Skill"
	KindAgent    = "Agent"
	KindPersona  = "Persona"
	KindProvider = "Provider"
	KindSchedule = "Schedule"
)

// TypeMeta is the apiVersion/kind envelope shared by every document. It is
// embedded inline so it serializes at the document's top level.
type TypeMeta struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
}

// Metadata is the shared metadata block. Name is required and, for most kinds,
// path-safe; Labels is an optional free-form string map.
type Metadata struct {
	Name   string            `json:"name" yaml:"name"`
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}
