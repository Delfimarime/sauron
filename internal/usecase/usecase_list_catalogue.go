package usecase

import (
	"context"
	"fmt"
	"io"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/internal/presentation"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/pkg/sauron/source"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// CatalogueKind is the kind of artifact a catalogue listing browses; it fixes
// the source root listed and the projection applied.
type CatalogueKind string

// The kinds a catalogue listing can browse.
const (
	// CatalogueSkill browses the skills a registry offers under .skills.
	CatalogueSkill CatalogueKind = "skill"
	// CatalogueAgent browses the agents a registry offers under .agents.
	CatalogueAgent CatalogueKind = "agent"
	// CataloguePersona browses the personas a registry offers under .personas.
	CataloguePersona CatalogueKind = "persona"
)

// the source roots holding each artifact kind's manifests.
const (
	rootSkills   = ".skills"
	rootAgents   = ".agents"
	rootPersonas = ".personas"
)

// catalogueRoots maps each kind to the source root holding its manifests.
var catalogueRoots = map[CatalogueKind]string{
	CatalogueSkill:   rootSkills,
	CatalogueAgent:   rootAgents,
	CataloguePersona: rootPersonas,
}

// the catalogue table column headers.
const (
	colName    = "NAME"
	colKind    = "KIND"
	colMembers = "MEMBERS"
)

// ListCatalogueUseCaseParams injects the collaborators the use case composes.
type ListCatalogueUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Open       OpenRegistry
	Logger     *zap.Logger
}

// ListCatalogueUseCase resolves a registry, opens its source live, and renders
// the artifacts of the selected kind followed by a paging line.
type ListCatalogueUseCase struct {
	registries storage.RegistriesStore
	open       OpenRegistry
	logger     *zap.Logger
}

// NewListCatalogueUseCase builds the use case from the injected collaborators.
func NewListCatalogueUseCase(params ListCatalogueUseCaseParams) *ListCatalogueUseCase {
	return &ListCatalogueUseCase{
		registries: params.Registries,
		open:       params.Open,
		logger:     params.Logger,
	}
}

// Execute runs the validate → find → open → list → project → render pipeline,
// returning a classified *Error on the first failing step.
func (uc *ListCatalogueUseCase) Execute(request *ListCatalogueRequest) error {
	if err := uc.validate(request); err != nil {
		return err
	}

	registry, err := uc.registries.FindByName(request.Context, request.Registry)
	if err != nil {
		return NewIOError(fmt.Sprintf("read registry %q: %v", request.Registry, err))
	}
	if registry == nil {
		return NewNotFoundError(fmt.Sprintf("registry %q does not exist", request.Registry))
	}

	fs, err := uc.open.Execute(request.Context, *registry)
	if err != nil {
		return err
	}

	files, err := uc.list(request, fs)
	if err != nil {
		return err
	}

	return uc.render(request, files)
}

// validate checks the inputs, returning a usage *Error for any out-of-range
// value. Sort and Order default to name and asc when unset.
func (uc *ListCatalogueUseCase) validate(request *ListCatalogueRequest) error {
	request.Sort, request.Order = defaultSortOrder(request.Sort, request.Order)
	if _, ok := catalogueRoots[request.Kind]; !ok {
		return NewUsageError(fmt.Sprintf("unknown kind %q", request.Kind))
	}
	if request.Sort != fieldName {
		return NewUsageError(fmt.Sprintf("unknown sort field %q", request.Sort))
	}
	if !isValidOrder(request.Order) {
		return NewUsageError(fmt.Sprintf("unknown order %q", request.Order))
	}
	if request.Page < 1 {
		return NewUsageError(fmt.Sprintf("page must be at least 1, got %d", request.Page))
	}
	if request.Limit < 1 {
		return NewUsageError(fmt.Sprintf("limit must be at least 1, got %d", request.Limit))
	}

	return nil
}

// list opens the source root for the kind and returns its entries, paging at the
// source with the computed offset.
func (uc *ListCatalogueUseCase) list(request *ListCatalogueRequest, fs source.FileSystem) ([]source.File, error) {
	opts := []source.Option{
		source.WithSort(fieldName),
		source.WithOrder(request.Order),
		source.WithOffset(request.offset()),
		source.WithLimit(request.Limit),
	}
	if request.Search != "" {
		opts = append([]source.Option{source.WithSearch(request.Search)}, opts...)
	}

	files, err := fs.List(request.Context, catalogueRoots[request.Kind], opts...)
	if err != nil {
		return nil, classifyAdapterErr(err)
	}

	return files, nil
}

// render builds the kind-specific rows, writes the table, then always writes the
// paging line, logging the outcome.
func (uc *ListCatalogueUseCase) render(request *ListCatalogueRequest, files []source.File) error {
	headers, rows, err := uc.project(request, files)
	if err != nil {
		return err
	}

	table := presentation.Table{Headers: headers, Rows: rows}
	if err := table.Render(request.Out()); err != nil {
		return NewIOError(fmt.Sprintf("render table: %v", err))
	}
	if _, err := fmt.Fprintln(request.Out(), uc.pagingLine(request, len(rows))); err != nil {
		return NewIOError(fmt.Sprintf("write paging line: %v", err))
	}

	uc.logger.Info("catalogue listed",
		zap.String(telemetry.FieldRegistryName, request.Registry),
		zap.Int(telemetry.FieldArtifactCount, len(rows)),
	)

	return nil
}

// project builds the headers and rows for the kind, skipping directory entries.
// Skills and agents render NAME/KIND with no content read; personas render
// NAME/MEMBERS by reading each manifest.
func (uc *ListCatalogueUseCase) project(request *ListCatalogueRequest, files []source.File) ([]string, [][]string, error) {
	if request.Kind == CataloguePersona {
		rows, err := uc.personaRows(request, files)
		if err != nil {
			return nil, nil, err
		}
		return []string{colName, colMembers}, rows, nil
	}

	manifests := filterBy(files, func(f source.File) bool { return !f.IsDirectory() })
	projectors := map[string]func(source.File) string{
		colName: func(f source.File) string { return catalogueName(f.Name()) },
		colKind: func(source.File) string { return string(request.Kind) },
	}
	rows := projectRows(manifests, []string{colName, colKind}, projectors)

	return []string{colName, colKind}, rows, nil
}

// personaRows reads each listed persona manifest and projects NAME/MEMBERS. A
// manifest that cannot be read or decoded is an io failure.
func (uc *ListCatalogueUseCase) personaRows(request *ListCatalogueRequest, files []source.File) ([][]string, error) {
	rows := make([][]string, 0, len(files))
	for _, file := range files {
		if file.IsDirectory() {
			continue
		}
		members, err := uc.readMembers(request.Context, file)
		if err != nil {
			return nil, err
		}
		rows = append(rows, []string{catalogueName(file.Name()), members})
	}

	return rows, nil
}

// readMembers reads and decodes a persona manifest, summarizing its declared
// skills and agents.
func (uc *ListCatalogueUseCase) readMembers(ctx context.Context, file source.File) (string, error) {
	reader, err := file.Read(ctx)
	if err != nil {
		return "", NewIOError(fmt.Sprintf("read persona %q: %v", file.Name(), err))
	}
	defer func() { _ = reader.Close() }()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", NewIOError(fmt.Sprintf("read persona %q: %v", file.Name(), err))
	}

	var persona types.Persona
	if err := yaml.Unmarshal(content, &persona); err != nil {
		return "", NewIOError(fmt.Sprintf("decode persona %q: %v", file.Name(), err))
	}

	return summarizeMembers(persona.Spec.Members), nil
}

// pagingLine renders the applied-paging report; an empty page reports zero
// results, a populated page the inclusive from–to window.
func (uc *ListCatalogueUseCase) pagingLine(request *ListCatalogueRequest, count int) string {
	if count == 0 {
		return fmt.Sprintf("showing 0 results (page %d, limit %d)", request.Page, request.Limit)
	}

	from := request.offset() + 1
	to := request.offset() + int64(count)

	return fmt.Sprintf("showing %d–%d (page %d, limit %d)", from, to, request.Page, request.Limit)
}

// catalogueName trims a manifest's .yaml or .yml extension to its catalogue name.
func catalogueName(filename string) string {
	for _, ext := range []string{".yaml", ".yml"} {
		if strings.HasSuffix(filename, ext) {
			return strings.TrimSuffix(filename, ext)
		}
	}

	return filename
}

// summarizeMembers renders a persona's membership as "skills: a, b; agents: c",
// omitting an empty group and rendering an em dash when both are empty.
func summarizeMembers(members types.PersonaMembers) string {
	var groups []string
	if len(members.Skills) > 0 {
		groups = append(groups, "skills: "+strings.Join(members.Skills, ", "))
	}
	if len(members.Agents) > 0 {
		groups = append(groups, "agents: "+strings.Join(members.Agents, ", "))
	}
	if len(groups) == 0 {
		return "—"
	}

	return strings.Join(groups, "; ")
}

// ListCatalogueRequest is the per-invocation input for browsing a registry's
// catalogue of one kind.
type ListCatalogueRequest struct {
	baseRequest

	Kind     CatalogueKind
	Registry string
	Search   string
	Sort     string
	Order    string
	Page     int64
	Limit    int64
}

// NewListCatalogueRequest builds a request bound to ctx and writing to out.
func NewListCatalogueRequest(ctx context.Context, out io.Writer) *ListCatalogueRequest {
	return &ListCatalogueRequest{baseRequest: baseRequest{Context: ctx, out: out}}
}

// offset translates the 1-based page and page size to a source offset.
func (r *ListCatalogueRequest) offset() int64 {
	return (r.Page - 1) * r.Limit
}
