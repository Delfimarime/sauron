package types

// Skill is an installed skill, recorded in track.yaml.
type Skill struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     ArtifactSpec `json:"spec" yaml:"spec"`
}

// ArtifactSpec is the spec block shared by Skill and Agent, whose schemas are
// identical. There is a single registry, so the source is implicit and is not
// recorded per artifact.
type ArtifactSpec struct {
	Version     string `json:"version,omitempty" yaml:"version,omitempty"`
	Digest      string `json:"digest" yaml:"digest"`
	Path        string `json:"path" yaml:"path"`
	InstalledAt string `json:"installedAt" yaml:"installedAt"`
	UpdatedAt   string `json:"updatedAt" yaml:"updatedAt"`
}
