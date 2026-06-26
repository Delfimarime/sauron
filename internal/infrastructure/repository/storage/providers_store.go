package storage

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/delfimarime/sauron/pkg/sauron/types"
)

// ProvidersStore is the typed view over the single Provider document. Sauron has
// exactly one provider, so the store is a singleton: it gets the one configured
// or sets it, replacing any already present — there is no name lookup.
type ProvidersStore interface {
	// Set stamps the document envelope and persists the provider, replacing any
	// provider already present.
	Set(ctx context.Context, p types.Provider) error
	// Get returns the configured provider, or nil when none is set.
	Get(ctx context.Context) (*types.Provider, error)
}

// providersStore implements ProvidersStore over a Store.
type providersStore struct {
	store *Store
}

// NewProvidersStore builds a ProvidersStore over store.
func NewProvidersStore(store *Store) ProvidersStore {
	return &providersStore{store: store}
}

// Get resolves the single provider document and decodes it into a Provider.
func (s *providersStore) Get(ctx context.Context) (*types.Provider, error) {
	node, err := s.store.First(ctx, types.KindProvider)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	var provider types.Provider
	if err := node.Decode(&provider); err != nil {
		return nil, fmt.Errorf("decode provider: %w", err)
	}

	return &provider, nil
}

// Set stamps the envelope, encodes the provider, and replaces the one present.
func (s *providersStore) Set(ctx context.Context, p types.Provider) error {
	p.TypeMeta = types.TypeMeta{
		APIVersion: types.APIVersion,
		Kind:       types.KindProvider,
	}

	var node yaml.Node
	if err := node.Encode(p); err != nil {
		return fmt.Errorf("encode provider: %w", err)
	}

	return s.store.Replace(ctx, types.KindProvider, &node)
}
