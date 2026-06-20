package types

// Persona is an installed persona and its last-resolved membership, recorded in
// track.yaml.
type Persona struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata    `json:"metadata" yaml:"metadata"`
	Spec     PersonaSpec `json:"spec" yaml:"spec"`
}

// PersonaSpec is the spec block of a Persona document.
type PersonaSpec struct {
	Registry    string         `json:"registry" yaml:"registry"`
	Version     string         `json:"version,omitempty" yaml:"version,omitempty"`
	Digest      string         `json:"digest" yaml:"digest"`
	Members     PersonaMembers `json:"members,omitempty" yaml:"members,omitempty"`
	InstalledAt string         `json:"installedAt" yaml:"installedAt"`
	UpdatedAt   string         `json:"updatedAt" yaml:"updatedAt"`
}

// PersonaMembers is the snapshot of last-resolved membership, for diffing on
// sync/upgrade.
type PersonaMembers struct {
	Skills []string `json:"skills,omitempty" yaml:"skills,omitempty"`
	Agents []string `json:"agents,omitempty" yaml:"agents,omitempty"`
}
