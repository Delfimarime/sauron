package storage

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// RegistriesStore is the typed view over the single Registry document. Sauron
// has exactly one registry, so the store is a singleton: it sets (replacing the
// one present), gets, or removes it — there is no name lookup and no listing.
type RegistriesStore interface {
	// Get returns the configured registry, or nil when none is set.
	Get(ctx context.Context) (*types.Registry, error)
	// Set stamps the document envelope and persists the registry, replacing any
	// registry already present.
	Set(ctx context.Context, r types.Registry) error
	// Remove drops the configured registry; an absent registry is a no-op.
	Remove(ctx context.Context) error
}

// registriesStore implements RegistriesStore over a Store.
type registriesStore struct {
	store *Store
}

// NewRegistriesStore builds a RegistriesStore over store.
func NewRegistriesStore(store *Store) RegistriesStore {
	return &registriesStore{store: store}
}

// Get resolves the single registry document and decodes it into a Registry.
func (s *registriesStore) Get(ctx context.Context) (*types.Registry, error) {
	node, err := s.store.First(ctx, types.KindRegistry)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	var registry types.Registry
	if err := node.Decode(&registry); err != nil {
		return nil, fmt.Errorf("decode registry: %w", err)
	}

	return &registry, nil
}

// Set stamps the envelope, encodes the registry, and replaces the one present.
func (s *registriesStore) Set(ctx context.Context, r types.Registry) error {
	r.TypeMeta = types.TypeMeta{
		APIVersion: types.APIVersion,
		Kind:       types.KindRegistry,
	}

	var node yaml.Node
	if err := node.Encode(r); err != nil {
		return fmt.Errorf("encode registry: %w", err)
	}

	return s.store.Replace(ctx, types.KindRegistry, &node)
}

// Remove drops the single registry document from the store.
func (s *registriesStore) Remove(ctx context.Context) error {
	return s.store.Purge(ctx, types.KindRegistry)
}
