package usecase

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// SetProviderUseCaseParams injects the stores and collaborators the use case
// composes.
type SetProviderUseCaseParams struct {
	fx.In
	Logger    *zap.Logger
	Providers storage.ProvidersStore
	Migrate   UseCase[MigrateInput, MigrateResult]
}

// SetProviderUseCase records the single global provider: it validates the name,
// no-ops when the provider is already active, and otherwise migrates the
// installed set to the new provider and persists it.
type SetProviderUseCase struct {
	logger    *zap.Logger
	providers storage.ProvidersStore
	migrate   UseCase[MigrateInput, MigrateResult]
}

// NewSetProviderUseCase builds the use case from the injected stores and
// collaborators.
func NewSetProviderUseCase(params SetProviderUseCaseParams) *SetProviderUseCase {
	return &SetProviderUseCase{
		providers: params.Providers,
		migrate:   params.Migrate,
		logger:    params.Logger,
	}
}

// Execute validates the requested provider, returns unchanged when it is already
// active, and otherwise migrates the installed set (on a real switch) and
// persists the provider — recording it even when some migration steps failed, so
// the setting and the track file stay consistent with what migrated.
func (uc *SetProviderUseCase) Execute(ctx context.Context, in SetProviderInput) (*SetProviderResult, error) {
	if err := uc.validateProviderName(in.Provider); err != nil {
		return nil, err
	}

	current, err := uc.providers.Get(ctx)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read provider: %v", err))
	}
	if current != nil && current.Metadata.Name == in.Provider {
		return &SetProviderResult{Provider: in.Provider, Unchanged: true}, nil
	}

	var moved []types.Artifact
	if current != nil {
		result, err := uc.migrate.Execute(ctx, MigrateInput{
			From: current.Metadata.Name, To: in.Provider,
		})
		if err != nil {
			return nil, err
		}
		moved = result.Moved
	}

	return uc.persist(ctx, in.Provider, moved)
}

// persist records the new provider with freshly stamped audit timestamps and
// projects the moved set into the presentation-agnostic plan groups.
func (uc *SetProviderUseCase) persist(ctx context.Context, name string, moved []types.Artifact) (*SetProviderResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	provider := types.Provider{Metadata: types.Metadata{
		Name:          name,
		CreatedAt:     now,
		LastUpdatedAt: now,
	}}
	if err := uc.providers.Set(ctx, provider); err != nil {
		return nil, NewIOError(fmt.Sprintf("persist provider: %v", err))
	}

	uc.logger.Debug("provider set", zap.String(telemetry.FieldProviderName, name))

	skills, agents := uc.groupByKind(moved)
	return &SetProviderResult{
		Provider: name,
		Skills:   skills,
		Agents:   agents,
		Migrated: len(moved),
	}, nil
}

// validateProviderName rejects any name Sauron does not support.
func (uc *SetProviderUseCase) validateProviderName(name string) error {
	switch name {
	case types.ProviderClaude, types.ProviderZencoder:
		return nil
	default:
		return NewUsageError(fmt.Sprintf("unknown provider %q", name))
	}
}

// groupByKind splits the moved artifacts into skill and agent name lists by their
// document kind.
func (uc *SetProviderUseCase) groupByKind(artifacts []types.Artifact) (skills, agents []string) {
	for _, artifact := range artifacts {
		switch artifact.Kind {
		case types.KindSkill:
			skills = append(skills, artifact.Metadata.Name)
		case types.KindAgent:
			agents = append(agents, artifact.Metadata.Name)
		}
	}
	return skills, agents
}

// SetProviderInput is the per-invocation input for setting the provider.
type SetProviderInput struct {
	Provider string
}

// SetProviderResult is the presentation-agnostic outcome of setting the
// provider: the provider now in effect, whether nothing changed, and the
// migration plan groups with their count.
type SetProviderResult struct {
	Migrated  int
	Unchanged bool
	Provider  string
	Skills    []string
	Agents    []string
}
