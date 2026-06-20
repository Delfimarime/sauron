package storage

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// RegistriesStore is the typed view over registry documents.
type RegistriesStore interface {
	// FindByName returns the registry with the given name, or nil when absent.
	FindByName(ctx context.Context, name string) (*types.Registry, error)
	// Add stamps the document envelope and persists the registry.
	Add(ctx context.Context, r types.Registry) error
}

// registriesStore implements RegistriesStore over a Store.
type registriesStore struct {
	store *Store
}

// NewRegistriesStore builds a RegistriesStore over store.
func NewRegistriesStore(store *Store) RegistriesStore {
	return &registriesStore{store: store}
}

// FindByName resolves a registry document and decodes it into a Registry.
func (s *registriesStore) FindByName(ctx context.Context, name string) (*types.Registry, error) {
	node, err := s.store.FindOne(ctx, types.KindRegistry, name)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	var registry types.Registry
	if err := node.Decode(&registry); err != nil {
		return nil, fmt.Errorf("decode registry %q: %w", name, err)
	}

	return &registry, nil
}

// Add stamps the envelope, encodes the registry, and appends it.
func (s *registriesStore) Add(ctx context.Context, r types.Registry) error {
	r.TypeMeta = types.TypeMeta{
		APIVersion: types.APIVersion,
		Kind:       types.KindRegistry,
	}

	var node yaml.Node
	if err := node.Encode(r); err != nil {
		return fmt.Errorf("encode registry %q: %w", r.Metadata.Name, err)
	}

	return s.store.Append(ctx, types.KindRegistry, &node)
}
