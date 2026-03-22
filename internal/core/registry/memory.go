package registry

import (
	"fmt"
	"sync"
)

type memoryRegistry struct {
	mu      sync.RWMutex
	queries map[string]string
	modules map[string][]string
}

func NewMemoryRegistry() Registry {
	return &memoryRegistry{
		queries: make(map[string]string),
		modules: make(map[string][]string),
	}
}

func (r *memoryRegistry) Register(module string, queries []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, q := range queries {
		if owner, exists := r.queries[q]; exists {
			return fmt.Errorf("registry: query %q already registered by %q", q, owner)
		}
	}

	for _, q := range queries {
		r.queries[q] = module
	}
	r.modules[module] = queries

	return nil
}

func (r *memoryRegistry) Lookup(query string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, exists := r.queries[query]
	if !exists {
		return "", ErrQueryNotFound
	}

	return module, nil
}

func (r *memoryRegistry) Deregister(module string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	queries, exists := r.modules[module]
	if !exists {
		return nil
	}

	for _, q := range queries {
		delete(r.queries, q)
	}
	delete(r.modules, module)

	return nil
}
