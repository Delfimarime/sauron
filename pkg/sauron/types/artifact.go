package types

// Artifact is the in-memory unification of an installed Skill or Agent. On disk
// each remains its own kind (Skill/Agent); in memory they share this shape and
// are discriminated by their embedded TypeMeta.Kind.
type Artifact struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     ArtifactSpec `json:"spec" yaml:"spec"`
}

// ArtifactSpec is the spec block shared by the Skill and Agent documents, whose
// schemas are identical. There is a single registry, so the source is implicit
// and is not recorded per artifact.
type ArtifactSpec struct {
	// Version is the identity read from the source — the artifact directory's git
	// tree hash, or the http object version — compared by sync and upgrade to
	// detect change.
	Version     string `json:"version" yaml:"version"`
	Path        string `json:"path" yaml:"path"`
	InstalledAt string `json:"installedAt" yaml:"installedAt"`
	UpdatedAt   string `json:"updatedAt" yaml:"updatedAt"`
}
