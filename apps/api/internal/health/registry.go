package health

import "context"

type Registry struct {
	checkers []Checker
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(c Checker) {
	r.checkers = append(r.checkers, c)
}

func (r *Registry) Check(ctx context.Context) map[string]error {

	results := make(map[string]error)

	for _, checker := range r.checkers {
		results[checker.Name()] = checker.Check(ctx)
	}

	return results
}
