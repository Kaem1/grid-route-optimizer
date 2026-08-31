package optimization

import "sort"

// Algorithm computes a Solution for a given Problem. Implementations must
// stay free of HTTP/JSON/rendering concerns so they can be tested and
// benchmarked in isolation (see project instructions, section 37).
type Algorithm interface {
	Name() string
	Solve(problem Problem) (Solution, error)
}

// Registry looks algorithms up by name so the API layer can select one
// requested by the frontend without depending on concrete implementations
// (see project instructions, section 38 — the project must support
// multiple algorithms).
type Registry struct {
	algorithms map[string]Algorithm
}

// NewRegistry builds a registry from the given algorithms, keyed by Name().
func NewRegistry(algorithms ...Algorithm) *Registry {
	r := &Registry{algorithms: make(map[string]Algorithm, len(algorithms))}
	for _, a := range algorithms {
		r.algorithms[a.Name()] = a
	}
	return r
}

// Get looks up an algorithm by name.
func (r *Registry) Get(name string) (Algorithm, bool) {
	a, ok := r.algorithms[name]
	return a, ok
}

// Names lists all registered algorithm names, sorted for stable output.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.algorithms))
	for name := range r.algorithms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
