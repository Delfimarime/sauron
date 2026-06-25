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

// ListCatalogueUseCaseParams injects the collaborators the use case composes.
type ListCatalogueUseCaseParams struct {
	fx.In
	Registries storage.RegistriesStore
	Open       OpenRegistry
	Logger     *zap.Logger
}

// ListCatalogueUseCase resolves a registry, opens its source live, and returns
// the artifacts of the selected kind.
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

// ListCatalogueInput is the input for browsing a registry's catalogue of one kind.
type ListCatalogueInput struct {
	Kind     CatalogueKind
	Registry string
	Search   string
	Sort     string
	Order    string
	Page     int64
	Limit    int64
}

// offset translates the 1-based page and page size to a source offset.
func (in ListCatalogueInput) offset() int64 {
	return (in.Page - 1) * in.Limit
}

// CatalogueEntry is one listed artifact: its catalogue name and, for personas,
// a membership summary (empty for skills and agents).
type CatalogueEntry struct {
	Name    string
	Members string
}

// ListCatalogueResult carries the listed entries and the applied paging window.
type ListCatalogueResult struct {
	Kind    CatalogueKind
	Entries []CatalogueEntry
	Page    int64
	Limit   int64
}

// Execute runs the validate → find → open → list → build pipeline, returning a
// classified *Error on the first failing step.
func (uc *ListCatalogueUseCase) Execute(ctx context.Context, in ListCatalogueInput) (*ListCatalogueResult, error) {
	in.Sort, in.Order = defaultSortOrder(in.Sort, in.Order)
	if err := uc.validate(in); err != nil {
		return nil, err
	}

	registry, err := uc.registries.FindByName(ctx, in.Registry)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read registry %q: %v", in.Registry, err))
	}
	if registry == nil {
		return nil, NewNotFoundError(fmt.Sprintf("registry %q does not exist", in.Registry))
	}

	fs, err := uc.open.Execute(ctx, *registry)
	if err != nil {
		return nil, err
	}

	files, err := uc.list(ctx, in, fs)
	if err != nil {
		return nil, err
	}

	entries, err := uc.entries(ctx, in, files)
	if err != nil {
		return nil, err
	}

	uc.logger.Info("catalogue listed",
		zap.String(telemetry.FieldRegistryName, in.Registry),
		zap.Int(telemetry.FieldArtifactCount, len(entries)),
	)

	return &ListCatalogueResult{Kind: in.Kind, Entries: entries, Page: in.Page, Limit: in.Limit}, nil
}

// validate checks the inputs, returning a usage *Error for any out-of-range
// value. Sort and Order are already defaulted to name and asc by the caller.
func (uc *ListCatalogueUseCase) validate(in ListCatalogueInput) error {
	if _, ok := catalogueRoots[in.Kind]; !ok {
		return NewUsageError(fmt.Sprintf("unknown kind %q", in.Kind))
	}
	if in.Sort != fieldName {
		return NewUsageError(fmt.Sprintf("unknown sort field %q", in.Sort))
	}
	if !isValidOrder(in.Order) {
		return NewUsageError(fmt.Sprintf("unknown order %q", in.Order))
	}
	if in.Page < 1 {
		return NewUsageError(fmt.Sprintf("page must be at least 1, got %d", in.Page))
	}
	if in.Limit < 1 {
		return NewUsageError(fmt.Sprintf("limit must be at least 1, got %d", in.Limit))
	}

	return nil
}

// list opens the source root for the kind and returns its entries, paging at the
// source with the computed offset.
func (uc *ListCatalogueUseCase) list(ctx context.Context, in ListCatalogueInput, fs source.FileSystem) ([]source.File, error) {
	opts := []source.Option{
		source.WithSort(fieldName),
		source.WithOrder(in.Order),
		source.WithOffset(in.offset()),
		source.WithLimit(in.Limit),
	}
	if in.Search != "" {
		opts = append([]source.Option{source.WithSearch(in.Search)}, opts...)
	}

	files, err := fs.List(ctx, catalogueRoots[in.Kind], opts...)
	if err != nil {
		return nil, classifyAdapterErr(err)
	}

	return files, nil
}

// entries projects the listed files into catalogue entries, skipping directory
// entries. Personas read each manifest to summarize membership; skills and
// agents carry the name alone.
func (uc *ListCatalogueUseCase) entries(ctx context.Context, in ListCatalogueInput, files []source.File) ([]CatalogueEntry, error) {
	entries := make([]CatalogueEntry, 0, len(files))
	for _, file := range files {
		if file.IsDirectory() {
			continue
		}
		entry := CatalogueEntry{Name: catalogueName(file.Name())}
		if in.Kind == CataloguePersona {
			members, err := uc.readMembers(ctx, file)
			if err != nil {
				return nil, err
			}
			entry.Members = members
		}
		entries = append(entries, entry)
	}

	return entries, nil
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
