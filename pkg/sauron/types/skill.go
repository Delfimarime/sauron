package types

// Skill is an installed skill, recorded in track.yaml.
type Skill struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     ArtifactSpec `json:"spec" yaml:"spec"`
}

// ArtifactSpec is the spec block shared by Skill and Agent, whose schemas are
// identical.
type ArtifactSpec struct {
	Registry    string     `json:"registry" yaml:"registry"`
	Version     string     `json:"version,omitempty" yaml:"version,omitempty"`
	Digest      string     `json:"digest" yaml:"digest"`
	Path        string     `json:"path" yaml:"path"`
	Provenance  Provenance `json:"provenance" yaml:"provenance"`
	InstalledAt string     `json:"installedAt" yaml:"installedAt"`
	UpdatedAt   string     `json:"updatedAt" yaml:"updatedAt"`
}

// Provenance records why an artifact is installed.
type Provenance struct {
	Direct   bool     `json:"direct" yaml:"direct"`
	Personas []string `json:"personas,omitempty" yaml:"personas,omitempty"`
}
