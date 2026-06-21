package usecase

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// DeleteArtifactsByRegistryResponse is the set of artifacts a cascade uninstall removes, grouped by kind.
// It is the value both delete-registry and uninstall report.
type DeleteArtifactsByRegistryResponse struct {
	Skills   []string
	Agents   []string
	Personas []string
}

// Total is the number of artifacts the plan removes across every kind.
func (p DeleteArtifactsByRegistryResponse) Total() int {
	return len(p.Skills) + len(p.Agents) + len(p.Personas)
}

// UninstallByRegistryActionParams injects the collaborators the action composes.
type UninstallByRegistryActionParams struct {
	fx.In
	Logger *zap.Logger
}

// UninstallByRegistryAction removes every tracked artifact sourced from a registry.
// It is the shared cleaning step both delete-registry and uninstall compose.
type UninstallByRegistryAction struct {
	logger *zap.Logger
}

// NewUninstallByRegistryAction builds the action from the injected collaborators.
func NewUninstallByRegistryAction(params UninstallByRegistryActionParams) *UninstallByRegistryAction {
	return &UninstallByRegistryAction{logger: params.Logger}
}

// Execute returns the plan of artifacts removed for the registry.
//
// 0007 owns the real body: the track store read, the per-artifact provider removal,
// and the track.yaml rewrite. Until then this is a no-op — it removes nothing and
// returns an empty plan, so an applied delete reports zero artifacts removed.
func (a *UninstallByRegistryAction) Execute(_ context.Context, _ string) (*DeleteArtifactsByRegistryResponse, error) {
	return &DeleteArtifactsByRegistryResponse{}, nil
}
