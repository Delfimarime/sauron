package usecase

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/presentation"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// ListRegistriesUseCaseParams injects the collaborators the use case composes.
type ListRegistriesUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Logger     *zap.Logger
}

// ListRegistriesUseCase reads, filters, sorts, and renders the registry listing.
type ListRegistriesUseCase struct {
	registries storage.RegistriesStore
	logger     *zap.Logger
}

// NewListRegistriesUseCase builds the use case from the injected collaborators.
func NewListRegistriesUseCase(params ListRegistriesUseCaseParams) *ListRegistriesUseCase {
	return &ListRegistriesUseCase{
		registries: params.Registries,
		logger:     params.Logger,
	}
}

// defaultColumns are the columns shown when --fields is not given.
func (uc *ListRegistriesUseCase) defaultColumns() []string {
	return []string{fieldName, fieldTransport, fieldURI}
}

// knownColumns is the set --fields may select from.
func (uc *ListRegistriesUseCase) knownColumns() map[string]struct{} {
	return map[string]struct{}{
		fieldName:      {},
		fieldTransport: {},
		fieldURI:       {},
		fieldRef:       {},
		fieldTimeout:   {},
	}
}

// sortColumns is the set --sort may select from.
func (uc *ListRegistriesUseCase) sortColumns() map[string]struct{} {
	return map[string]struct{}{
		fieldName:      {},
		fieldTransport: {},
	}
}

// projectors maps each column to the registry field it reads.
func (uc *ListRegistriesUseCase) projectors() map[string]func(types.Registry) string {
	return map[string]func(types.Registry) string{
		fieldName:      func(r types.Registry) string { return r.Metadata.Name },
		fieldTransport: func(r types.Registry) string { return string(r.Spec.Transport) },
		fieldURI:       func(r types.Registry) string { return r.Spec.URI },
		fieldRef:       func(r types.Registry) string { return r.Spec.Ref },
		fieldTimeout:   func(r types.Registry) string { return r.Spec.Timeout },
	}
}

// Execute runs the read → filter → sort → project → render pipeline, returning a
// *Error on the first failing step.
func (uc *ListRegistriesUseCase) Execute(request *ListRegistriesRequest) error {
	fields, err := uc.determineFields(request.Fields)
	if err != nil {
		return err
	}
	sortBy, order, err := uc.determineOrder(request.Sort, request.Order)
	if err != nil {
		return err
	}

	registries, err := uc.registries.List(request.Context)
	if err != nil {
		return NewIOError(fmt.Sprintf("read registries: %v", err))
	}

	registries = filterBy(registries, uc.nameContains(request.Search))
	uc.orderRegistries(registries, sortBy, order)

	return uc.render(request, registries, fields)
}

// render projects the registries onto the selected columns and writes the table,
// logging the outcome. An empty listing skips the renderer and writes nothing.
func (uc *ListRegistriesUseCase) render(request *ListRegistriesRequest, registries []types.Registry, fields []string) error {
	uc.logger.Info("registries listed",
		zap.Int(telemetry.FieldRegistryCount, len(registries)),
	)
	if len(registries) == 0 {
		return nil
	}

	table := presentation.Table{Headers: fields, Rows: uc.rows(registries, fields)}
	if err := table.Render(request.Out()); err != nil {
		return NewIOError(fmt.Sprintf("render table: %v", err))
	}

	return nil
}

// determineFields validates the requested columns and forces name present and
// first; an empty request yields the default columns.
func (uc *ListRegistriesUseCase) determineFields(requested []string) ([]string, error) {
	return selectFields(requested, uc.knownColumns(), uc.defaultColumns())
}

// determineOrder validates the sort field and direction, applying the defaults.
func (uc *ListRegistriesUseCase) determineOrder(sortBy, order string) (string, string, error) {
	sortBy, order = defaultSortOrder(sortBy, order)
	if _, ok := uc.sortColumns()[sortBy]; !ok {
		return "", "", NewUsageError(fmt.Sprintf("unknown sort field %q", sortBy))
	}
	if !isValidOrder(order) {
		return "", "", NewUsageError(fmt.Sprintf("unknown order %q", order))
	}

	return sortBy, order, nil
}

// nameContains matches registries whose name contains the term,
// case-insensitively; an empty term matches every registry.
func (uc *ListRegistriesUseCase) nameContains(search string) predicate[types.Registry] {
	term := strings.ToLower(search)
	return func(r types.Registry) bool {
		return strings.Contains(strings.ToLower(r.Metadata.Name), term)
	}
}

// orderRegistries sorts the registries by the sort field and direction, always
// breaking ties on name ascending for a deterministic order.
func (uc *ListRegistriesUseCase) orderRegistries(registries []types.Registry, sortBy, order string) {
	key := uc.sortKey(sortBy)
	slices.SortStableFunc(registries, func(a, b types.Registry) int {
		primary := strings.Compare(key(a), key(b))
		if order == orderDesc {
			primary = -primary
		}
		return cmp.Or(primary, strings.Compare(a.Metadata.Name, b.Metadata.Name))
	})
}

// sortKey maps a registry to its comparison key for the sort field.
func (uc *ListRegistriesUseCase) sortKey(sortBy string) func(types.Registry) string {
	if sortBy == fieldTransport {
		return func(r types.Registry) string { return string(r.Spec.Transport) }
	}
	return func(r types.Registry) string { return r.Metadata.Name }
}

// rows projects each registry onto the selected columns.
func (uc *ListRegistriesUseCase) rows(registries []types.Registry, fields []string) [][]string {
	return projectRows(registries, fields, uc.projectors())
}

// ListRegistriesRequest is the per-invocation input for the registry listing.
type ListRegistriesRequest struct {
	baseRequest

	Search string
	Fields []string
	Sort   string
	Order  string
}

// NewListRegistriesRequest builds a request bound to ctx and writing to out.
func NewListRegistriesRequest(ctx context.Context, out io.Writer) *ListRegistriesRequest {
	return &ListRegistriesRequest{baseRequest: baseRequest{Context: ctx, out: out}}
}
