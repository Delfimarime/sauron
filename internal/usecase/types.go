package usecase

import (
	"time"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// -- Diff types --

// DesiredArtifact is one artifact a caller wants present at a specific version,
// identified by its kind and name. It carries no content.
type DesiredArtifact struct {
	Kind    string
	Name    string
	Version string
}

// UpdatePlan pairs the tracked artifact that will be replaced with the desired
// version to install. Prior carries the recorded artifact so downstream steps
// preserve its InstalledAt without re-reading the track store.
type UpdatePlan struct {
	Prior   types.Artifact
	Desired DesiredArtifact
}

// DiffRequest is the per-invocation input: the desired artifacts and whether
// tracked artifacts absent from the desired set should be reported for removal.
type DiffRequest struct {
	Desired         []DesiredArtifact
	IncludeRemovals bool
}

// DiffResponse is the presentation-agnostic reconciliation plan: the artifacts to
// add (as desired entries), the artifacts to update (each carrying the prior
// tracked artifact alongside the desired version), the tracked artifacts to remove
// (only when requested), and the tracked artifacts that are already current.
type DiffResponse struct {
	Add       []DesiredArtifact
	Update    []UpdatePlan
	Remove    []types.Artifact
	Unchanged []types.Artifact
}

// -- Install types --

// InstallRequest is the per-invocation input: the artifact Kind (Skill or Agent)
// and the names to install.
type InstallRequest struct {
	Kind  string
	Names []string
}

// InstallResponse is the presentation-agnostic outcome of an install: the
// artifacts added and updated (each carrying its Kind) and the per-name failures.
type InstallResponse struct {
	Added    []types.Artifact
	Updated  []types.Artifact
	Failures []InstallFailure
}

// skip records a benign per-name skip — a name the registry does not offer
// (FR-006) or one whose source declares no version — leaving siblings to install.
// A skip does not fail the install: the command still exits 0.
func (r *InstallResponse) skip(name, reason string) {
	r.Failures = append(r.Failures, InstallFailure{Name: name, Reason: reason})
}

// fail records a per-name persist failure — a fetch, write, or track update that
// failed — leaving siblings to install (FR-006). A persist failure means the
// install could not be persisted: the command exits non-zero.
func (r *InstallResponse) fail(name, reason string) {
	r.Failures = append(r.Failures, InstallFailure{Name: name, Reason: reason, Fatal: true})
}

// InstallFailure records one name that could not be installed and why.
// It is distinct from MigrateFailure: install keys failures by artifact name
// (a string) whereas migration keys them by the full tracked artifact.
//
// Fatal distinguishes a persist failure (fetch, write, or track update on an
// offered artifact — the install could not be persisted, exit 1) from a benign
// skip (a name not offered or with no declared version — exit 0).
type InstallFailure struct {
	Name   string
	Reason string
	Fatal  bool
}

// -- Migrate types --

// MigrateRequest names the source and destination providers by name.
type MigrateRequest struct {
	From string
	To   string
}

// MigrateResponse is the presentation-agnostic outcome of a migration: the
// artifacts moved (each carrying its Kind) and the per-artifact failures.
type MigrateResponse struct {
	Moved    []types.Artifact
	Failures []MigrateFailure
}

// MigrateFailure records one artifact that could not be migrated and why.
// It is distinct from InstallFailure: migration keys failures by the full tracked
// artifact (needed for logging and correlation) whereas install keys by name only.
type MigrateFailure struct {
	Reason   string
	Artifact types.Artifact
}

// -- List Catalogue types --

// CatalogueKind is the kind of artifact a catalogue listing browses; it fixes
// the source root listed and the projection applied.
type CatalogueKind string

// The kinds a catalogue listing can browse.
const (
	// CatalogueSkill browses the skills the registry offers under skills.
	CatalogueSkill CatalogueKind = "skill"
	// CatalogueAgent browses the agents the registry offers under agents.
	CatalogueAgent CatalogueKind = "agent"
)

// ListCatalogueRequest is the per-invocation input for browsing the registry's
// catalogue of one kind.
type ListCatalogueRequest struct {
	Kind   CatalogueKind
	Search string
	Sort   string
	Order  string
	Page   int64
	Limit  int64
}

// offset translates the 1-based page and page size to a source offset.
func (in ListCatalogueRequest) offset() int64 {
	return (in.Page - 1) * in.Limit
}

// ListCatalogueResponse is the presentation-agnostic outcome of browsing the
// catalogue: the artifact names of the kind and the paging window applied.
type ListCatalogueResponse struct {
	Kind   CatalogueKind
	Items  []string
	Page   int64
	Limit  int64
	Offset int64
}

// -- Set Registry types --

// SetRegistryRequest is the per-invocation input for configuring the registry.
type SetRegistryRequest struct {
	Source    string
	Transport string
	Revision  string
	Username  string
	Password  string
	SSHKey    string

	SkipTLSVerify bool
	CACert        string
	ClientCert    string
	ClientKey     string

	Timeout time.Duration
}

// SetRegistryResponse is the presentation-agnostic outcome of configuring the
// registry: the source now in effect and the transport it is reached over.
type SetRegistryResponse struct {
	Source    string
	Transport types.Transport
}

// -- Set Provider types --

// SetProviderRequest is the per-invocation input for setting the provider.
type SetProviderRequest struct {
	Provider string
}

// SetProviderResponse is the presentation-agnostic outcome of setting the
// provider: the provider now in effect, whether nothing changed, the migration
// plan groups with their count, and any artifacts the migration could not move
// (stranded under the old provider directory).
type SetProviderResponse struct {
	Migrated  int
	Unchanged bool
	Provider  string
	Skills    []string
	Agents    []string
	Failures  []MigrateFailure
}

// -- Describe Registry types --

// DescribeRegistryRequest is the per-invocation input for describing the registry.
// Describing the single configured registry takes no business input; field
// selection is a presentation concern resolved by the caller.
type DescribeRegistryRequest struct{}

// -- Describe Provider types --

// DescribeProviderRequest is the per-invocation input for describing the provider.
// Describing the single configured provider takes no business input; field
// selection is a presentation concern resolved by the caller.
type DescribeProviderRequest struct{}

// -- Unset Registry types --

// UnsetOutcome classifies which removal outcome occurred, so the client can
// render the matching report.
type UnsetOutcome string

// The outcomes unsetting the registry can produce.
const (
	// UnsetNothing reports no registry was configured, so nothing was unset.
	UnsetNothing UnsetOutcome = "nothing"
	// UnsetPreview reports a dry-run preview that changed no state.
	UnsetPreview UnsetOutcome = "preview"
	// UnsetRemoved reports the configured registry was removed.
	UnsetRemoved UnsetOutcome = "removed"
)

// UnsetRegistryRequest is the per-invocation input for removing the registry.
type UnsetRegistryRequest struct {
	DryRun bool
}

// UnsetRegistryResponse is the presentation-agnostic outcome of unsetting.
type UnsetRegistryResponse struct {
	Outcome UnsetOutcome
}
