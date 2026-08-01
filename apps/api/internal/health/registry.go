package health

import "context"

type Registry struct {
	checkers []Checker
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(checker Checker) {
	if checker == nil {
		return
	}

	r.checkers = append(r.checkers, checker)
}

func (r *Registry) Check(ctx context.Context) map[string]error {
	results := make(map[string]error, len(r.checkers))

	for _, checker := range r.checkers {
		results[checker.Name()] = checker.Check(ctx)
	}

	return results
}
