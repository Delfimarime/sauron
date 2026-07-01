package usecase

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"github.com/delfimarime/sauron/internal/infrastructure/repository/storage"
	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// DiffUseCaseParams injects the store the diff reconciles against.
type DiffUseCaseParams struct {
	fx.In
	Track storage.TrackStore
}

// DiffUseCase computes the install/sync/upgrade plan purely on versions: it reads
// the recorded set, indexes it by (kind, name), and categorizes each desired
// artifact as an addition (untracked), an update (tracked, version differs), or
// unchanged (tracked, version equal). When removals are requested it also reports
// the tracked artifacts absent from the desired set. It reads no artifact content,
// so install, sync, and upgrade reuse it. It fits the generic
// UseCase[DiffRequest, DiffResponse] shape and never renders.
type DiffUseCase struct {
	track storage.TrackStore
}

// NewDiffUseCase builds the use case from the injected store.
func NewDiffUseCase(params DiffUseCaseParams) *DiffUseCase {
	return &DiffUseCase{track: params.Track}
}

// Execute reads the recorded set and categorizes every desired artifact on its
// version, appending removals only when the input requests them.
func (uc *DiffUseCase) Execute(ctx context.Context, in DiffRequest) (*DiffResponse, error) {
	tracked, err := uc.track.List(ctx)
	if err != nil {
		return nil, NewIOError(fmt.Sprintf("read installed set: %v", err))
	}

	index := make(map[string]types.Artifact, len(tracked))
	for _, artifact := range tracked {
		index[trackKey(artifact.Kind, artifact.Metadata.Name)] = artifact
	}

	diff := &DiffResponse{}
	desired := make(map[string]struct{}, len(in.Desired))
	for _, want := range in.Desired {
		desired[trackKey(want.Kind, want.Name)] = struct{}{}
		diff.categorize(want, index)
	}

	if in.IncludeRemovals {
		diff.removals(tracked, desired)
	}

	return diff, nil
}

// categorize sorts one desired artifact into Add (untracked), Update (tracked and
// the version differs), or Unchanged (tracked and the version is equal). For an
// update, the prior tracked artifact is carried in UpdatePlan so downstream steps
// can preserve its InstalledAt without re-reading the track store.
func (d *DiffResponse) categorize(want DesiredArtifact, index map[string]types.Artifact) {
	prior, tracked := index[trackKey(want.Kind, want.Name)]
	switch {
	case !tracked:
		d.Add = append(d.Add, want)
	case prior.Spec.Version == want.Version:
		d.Unchanged = append(d.Unchanged, prior)
	default:
		d.Update = append(d.Update, UpdatePlan{Prior: prior, Desired: want})
	}
}

// removals appends every tracked artifact whose key is absent from the desired
// set, leaving untouched siblings of a partial desired set in place.
func (d *DiffResponse) removals(tracked []types.Artifact, desired map[string]struct{}) {
	for _, artifact := range tracked {
		if _, want := desired[trackKey(artifact.Kind, artifact.Metadata.Name)]; !want {
			d.Remove = append(d.Remove, artifact)
		}
	}
}
