package engine

import (
	"strings"
	"sync"
)

// Registry manages all registered search engine instances.
type Registry struct {
	mu         sync.RWMutex
	engines    map[string]Engine
	shortcuts  map[string]string
	categories map[string][]string
}

// GlobalRegistry is the default shared engine registry.
var GlobalRegistry = NewRegistry()

// NewRegistry creates a new engine registry.
func NewRegistry() *Registry {
	return &Registry{
		engines:    make(map[string]Engine),
		shortcuts:  make(map[string]string),
		categories: make(map[string][]string),
	}
}

// Register registers an engine implementation in the registry.
func (r *Registry) Register(e Engine) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := strings.ToLower(e.Name())
	r.engines[name] = e

	for _, cat := range e.Categories() {
		catLower := strings.ToLower(cat)
		r.categories[catLower] = append(r.categories[catLower], name)
	}
}

// RegisterShortcut associates a bang/shortcut like "!g" with an engine name.
func (r *Registry) RegisterShortcut(shortcut, engineName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shortcuts[strings.ToLower(shortcut)] = strings.ToLower(engineName)
}

// Get returns an engine by its name.
func (r *Registry) Get(name string) (Engine, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.engines[strings.ToLower(name)]
	return e, ok
}

// GetByShortcut returns an engine mapped to a shortcut bang.
func (r *Registry) GetByShortcut(shortcut string) (Engine, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	engineName, ok := r.shortcuts[strings.ToLower(shortcut)]
	if !ok {
		return nil, false
	}
	e, found := r.engines[engineName]
	return e, found
}

// GetByCategory returns all engine instances belonging to a category.
func (r *Registry) GetByCategory(category string) []Engine {
	r.mu.RLock()
	defer r.mu.RUnlock()

	engineNames, ok := r.categories[strings.ToLower(category)]
	if !ok {
		return nil
	}

	result := make([]Engine, 0, len(engineNames))
	for _, name := range engineNames {
		if e, found := r.engines[name]; found {
			result = append(result, e)
		}
	}
	return result
}

// GetAll returns all registered engine instances.
func (r *Registry) GetAll() []Engine {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Engine, 0, len(r.engines))
	for _, e := range r.engines {
		result = append(result, e)
	}
	return result
}

// GetEngineWeights returns a map of all engine weights.
func (r *Registry) GetEngineWeights() map[string]float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	weights := make(map[string]float64, len(r.engines))
	for name, e := range r.engines {
		weights[name] = e.Weight()
	}
	return weights
}
