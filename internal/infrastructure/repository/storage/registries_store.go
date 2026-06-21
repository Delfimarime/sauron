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
	// List returns every stored registry, validated on read.
	List(ctx context.Context) ([]types.Registry, error)
	// Remove drops the registry with the given name; an absent name is a no-op.
	Remove(ctx context.Context, name string) error
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

// Remove drops the named registry document from the store.
func (s *registriesStore) Remove(ctx context.Context, name string) error {
	return s.store.Remove(ctx, types.KindRegistry, name)
}

// List reads every registry document and decodes each into a Registry.
func (s *registriesStore) List(ctx context.Context) ([]types.Registry, error) {
	nodes, err := s.store.FindAll(ctx, types.KindRegistry)
	if err != nil {
		return nil, err
	}

	registries := make([]types.Registry, 0, len(nodes))
	for _, node := range nodes {
		var registry types.Registry
		if err := node.Decode(&registry); err != nil {
			return nil, fmt.Errorf("decode registry: %w", err)
		}
		registries = append(registries, registry)
	}

	return registries, nil
}
