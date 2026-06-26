package storage

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// TrackStore is the view over the installed set recorded in track.yaml. It reads
// the Skill and Agent documents as a single Artifact slice (discriminated by the
// embedded TypeMeta.Kind) and updates one artifact in place.
type TrackStore interface {
	// List returns every recorded artifact, each carrying its document Kind, or
	// an empty slice when track.yaml is absent.
	List(ctx context.Context) ([]types.Artifact, error)
	// Update persists artifact, replacing the matching document (routed by its
	// Kind) or appending it when absent.
	Update(ctx context.Context, artifact types.Artifact) error
}

// trackStore implements TrackStore over a Store.
type trackStore struct {
	store *Store
}

// NewTrackStore builds a TrackStore over store.
func NewTrackStore(store *Store) TrackStore {
	return &trackStore{store: store}
}

// List reads the Skill and Agent documents from track.yaml and decodes them into
// a single Artifact slice, each stamped with its document kind envelope.
func (s *trackStore) List(ctx context.Context) ([]types.Artifact, error) {
	skills, err := s.collect(ctx, types.KindSkill)
	if err != nil {
		return nil, err
	}

	agents, err := s.collect(ctx, types.KindAgent)
	if err != nil {
		return nil, err
	}

	return append(skills, agents...), nil
}

// collect reads every document of kind and decodes it into an Artifact stamped
// with the kind envelope, so callers can discriminate by Artifact.Kind.
func (s *trackStore) collect(ctx context.Context, kind string) ([]types.Artifact, error) {
	docs, err := s.store.FindAll(ctx, kind)
	if err != nil {
		return nil, err
	}

	artifacts := make([]types.Artifact, 0, len(docs))
	for _, doc := range docs {
		var artifact types.Artifact
		if err := doc.Decode(&artifact); err != nil {
			return nil, fmt.Errorf("decode %s: %w", kind, err)
		}
		artifact.TypeMeta = types.TypeMeta{APIVersion: types.APIVersion, Kind: kind}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// Update routes the artifact to its document kind, stamps the envelope, and
// upserts the document.
func (s *trackStore) Update(ctx context.Context, artifact types.Artifact) error {
	kind := artifact.Kind
	name := artifact.Metadata.Name
	artifact.TypeMeta = types.TypeMeta{APIVersion: types.APIVersion, Kind: kind}

	var node yaml.Node
	if err := node.Encode(artifact); err != nil {
		return fmt.Errorf("encode artifact: %w", err)
	}

	return s.store.Upsert(ctx, kind, name, &node)
}
