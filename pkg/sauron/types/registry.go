package types

// Transport identifies how a registry's source is reached. Mirrors the
// spec.transport enum in spec/contracts/schemas/Registry.schema.json.
type Transport string

// The supported registry transports.
const (
	TransportGit        Transport = "git"
	TransportHTTP       Transport = "http"
	TransportFilesystem Transport = "filesystem"
)

// Registry is a registered source of artifacts, persisted in registries.yaml.
// Mirrors spec/contracts/schemas/Registry.schema.json.
type Registry struct {
	TypeMeta `json:",inline" yaml:",inline"`
	Metadata Metadata     `json:"metadata" yaml:"metadata"`
	Spec     RegistrySpec `json:"spec" yaml:"spec"`
}

// RegistrySpec is the spec block of a Registry document.
type RegistrySpec struct {
	Transport Transport `json:"transport" yaml:"transport"`
	URI       string    `json:"uri" yaml:"uri"`
	Auth      *Auth     `json:"auth,omitempty" yaml:"auth,omitempty"`
	TLS       *TLS      `json:"tls,omitempty" yaml:"tls,omitempty"`
	SSHKey    string    `json:"sshKey,omitempty" yaml:"sshKey,omitempty"`
	// Timeout is a Go duration string bounding network operations (default 30s).
	Timeout string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// Auth holds credentials as environment references only (${env:VAR}); the
// values are stored verbatim and never resolved into secrets.
type Auth struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

// TLS holds optional transport-security settings for the http transport.
type TLS struct {
	SkipVerify bool   `json:"skipVerify,omitempty" yaml:"skipVerify,omitempty"`
	CACert     string `json:"caCert,omitempty" yaml:"caCert,omitempty"`
	ClientCert string `json:"clientCert,omitempty" yaml:"clientCert,omitempty"`
	ClientKey  string `json:"clientKey,omitempty" yaml:"clientKey,omitempty"`
}
