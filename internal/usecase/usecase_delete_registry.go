package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/telemetry"
)

// DeleteRegistryUseCaseParams injects the collaborators the use case composes.
type DeleteRegistryUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Cascade    *UninstallByRegistryAction
	Logger     *zap.Logger
}

// DeleteRegistryUseCase unregisters a source and cascade-uninstalls its artifacts.
type DeleteRegistryUseCase struct {
	registries storage.RegistriesStore
	cascade    *UninstallByRegistryAction
	logger     *zap.Logger
}

// NewDeleteRegistryUseCase builds the use case from the injected collaborators.
func NewDeleteRegistryUseCase(params DeleteRegistryUseCaseParams) *DeleteRegistryUseCase {
	return &DeleteRegistryUseCase{
		registries: params.Registries,
		cascade:    params.Cascade,
		logger:     params.Logger,
	}
}

// Execute runs the find → cascade → dry-run → remove → report pipeline, returning a
// *Error on the first failing step. A registry that does not exist is a success.
func (uc *DeleteRegistryUseCase) Execute(request *DeleteRegistryRequest) error {
	registry, err := uc.registries.FindByName(request.Context, request.Name)
	if err != nil {
		return NewIOError(fmt.Sprintf("lookup registry %q: %v", request.Name, err))
	}
	if registry == nil {
		return uc.reportNothingDeleted(request)
	}

	plan, err := uc.cascade.Execute(request.Context, request.Name)
	if err != nil {
		return NewIOError(fmt.Sprintf("uninstall artifacts of %q: %v", request.Name, err))
	}

	if request.DryRun {
		return uc.reportDryRun(request, plan)
	}

	if err := uc.registries.Remove(request.Context, request.Name); err != nil {
		return NewIOError(fmt.Sprintf("remove registry %q: %v", request.Name, err))
	}

	return uc.reportRemoved(request, plan)
}

// reportNothingDeleted reports the FR-005 idempotent-delete outcome: nothing of that
// name existed, so nothing was deleted, and the command still succeeds.
func (uc *DeleteRegistryUseCase) reportNothingDeleted(request *DeleteRegistryRequest) error {
	uc.logger.Debug("registry not found",
		zap.String(telemetry.FieldRegistryName, request.Name),
	)
	return uc.write(request, fmt.Sprintf("registry %q does not exist; nothing was deleted\n", request.Name))
}

// reportDryRun previews the cascade plan without changing any state.
func (uc *DeleteRegistryUseCase) reportDryRun(request *DeleteRegistryRequest, plan *DeleteArtifactsByRegistryResponse) error {
	uc.logger.Debug("registry deletion previewed",
		zap.String(telemetry.FieldRegistryName, request.Name),
		zap.Int(telemetry.FieldArtifactCount, plan.Total()),
	)
	summary := fmt.Sprintf("registry %q would be removed; %d artifacts would be removed\n", request.Name, plan.Total())
	return uc.write(request, uc.renderPlan(plan)+summary)
}

// reportRemoved reports the applied removal: the grouped plan followed by the
// summary count.
func (uc *DeleteRegistryUseCase) reportRemoved(request *DeleteRegistryRequest, plan *DeleteArtifactsByRegistryResponse) error {
	uc.logger.Debug("registry removed",
		zap.String(telemetry.FieldRegistryName, request.Name),
		zap.Int(telemetry.FieldArtifactCount, plan.Total()),
	)
	summary := fmt.Sprintf("registry %q removed; %d artifacts removed\n", request.Name, plan.Total())
	return uc.write(request, uc.renderPlan(plan)+summary)
}

// renderPlan formats the grouped removal plan, printing only the non-empty kind
// groups; each group is a heading with a "-"-prefixed entry per artifact.
func (uc *DeleteRegistryUseCase) renderPlan(plan *DeleteArtifactsByRegistryResponse) string {
	var b strings.Builder
	uc.renderGroup(&b, "skills", plan.Skills)
	uc.renderGroup(&b, "agents", plan.Agents)
	uc.renderGroup(&b, "personas", plan.Personas)
	return b.String()
}

// renderGroup renders one kind heading and its entries into b, or nothing when empty.
func (uc *DeleteRegistryUseCase) renderGroup(b *strings.Builder, heading string, names []string) {
	if len(names) == 0 {
		return
	}
	b.WriteString(heading)
	b.WriteString(":\n")
	for _, name := range names {
		b.WriteString("  - ")
		b.WriteString(name)
		b.WriteString("\n")
	}
}

// write emits text to the request's output, classifying a write failure as io.
func (uc *DeleteRegistryUseCase) write(request *DeleteRegistryRequest, text string) error {
	if _, err := fmt.Fprint(request.Out(), text); err != nil {
		return NewIOError(fmt.Sprintf("write report: %v", err))
	}
	return nil
}

// DeleteRegistryRequest is the per-invocation input for deleting a source.
type DeleteRegistryRequest struct {
	context.Context
	out io.Writer

	Name   string
	DryRun bool
}

// NewDeleteRegistryRequest builds a request bound to ctx and writing to out.
func NewDeleteRegistryRequest(ctx context.Context, out io.Writer) *DeleteRegistryRequest {
	return &DeleteRegistryRequest{Context: ctx, out: out}
}

// Out returns the writer the command's output goes to.
func (r *DeleteRegistryRequest) Out() io.Writer {
	return r.out
}
