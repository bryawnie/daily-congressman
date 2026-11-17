// Package source holds the registry of chamber data sources.
package source

import (
	"fmt"
	"sort"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// Registry maps each chamber to its Source.
type Registry struct {
	sources map[congress.Chamber]congress.Source
}

// NewRegistry builds a registry from the given sources.
func NewRegistry(sources ...congress.Source) *Registry {
	r := &Registry{sources: make(map[congress.Chamber]congress.Source)}
	for _, s := range sources {
		r.sources[s.Chamber()] = s
	}
	return r
}

// Get returns the source of a chamber.
func (r *Registry) Get(c congress.Chamber) (congress.Source, error) {
	s, ok := r.sources[c]
	if !ok {
		return nil, fmt.Errorf("source: no source registered for chamber %q", c)
	}
	return s, nil
}

// Chambers returns the registered chambers in a stable order.
func (r *Registry) Chambers() []congress.Chamber {
	out := make([]congress.Chamber, 0, len(r.sources))
	for c := range r.sources {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Only returns a registry restricted to the given chambers. It fails if one of
// them is not registered, so a typo in the configuration is not silently ignored.
func (r *Registry) Only(chambers ...congress.Chamber) (*Registry, error) {
	out := &Registry{sources: make(map[congress.Chamber]congress.Source)}
	for _, c := range chambers {
		s, err := r.Get(c)
		if err != nil {
			return nil, err
		}
		out.sources[c] = s
	}
	return out, nil
}
