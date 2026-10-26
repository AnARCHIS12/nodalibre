package provider

import (
	"fmt"
	"sort"
	"sync"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]ChatProvider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]ChatProvider)}
}

func (r *Registry) Register(p ChatProvider) error {
	if p == nil || p.ID() == "" {
		return fmt.Errorf("provider id is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.providers[p.ID()]; exists {
		return fmt.Errorf("provider %q is already registered", p.ID())
	}
	r.providers[p.ID()] = p
	return nil
}

func (r *Registry) Get(id string) (ChatProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[id]
	return p, ok
}

func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
