package types

// The recognized provider destinations (Provider.metadata.name enum).
const (
	ProviderClaude   = "claude"
	ProviderZencoder = "zencoder"
)

// Provider is the single global provider destination, recorded in
// settings.yaml. It has no spec — its identity is its name.
type Provider struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata `json:"metadata" yaml:"metadata"`
}
