package types

// The recognized provider destinations (Provider.metadata.name enum).
const (
	ProviderClaude   = "claude"
	ProviderZencoder = "zencoder"
)

// Provider is the single global provider destination, recorded in
// settings.yaml. Its identity is its name; its spec carries the sync timestamps.
type Provider struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     ProviderSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

// ProviderSpec carries the provider's sync state. Both timestamps are RFC3339 UTC
// stamps written by sync (0011) and only read elsewhere; they are tolerated absent
// on read — a provider that has never synced carries neither.
type ProviderSpec struct {
	LastSyncedAt      string `json:"lastSyncedAt,omitempty" yaml:"lastSyncedAt,omitempty"`
	LastSyncAttemptAt string `json:"lastSyncAttemptAt,omitempty" yaml:"lastSyncAttemptAt,omitempty"`
}
