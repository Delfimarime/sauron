package types

// Agent is an installed agent, recorded in track.yaml. Its spec shape is
// identical to Skill (see ArtifactSpec).
type Agent struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     ArtifactSpec `json:"spec" yaml:"spec"`
}
