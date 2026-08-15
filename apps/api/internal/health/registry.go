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

func (r *Registry) Check(ctx context.Context) []Result {
	results := make([]Result, 0, len(r.checkers))

	for _, checker := range r.checkers {
		err := checker.Check(ctx)

		result := Result{
			Name: checker.Name(),
		}

		if err != nil {
			result.Error = err.Error()
		}

		results = append(results, result)
	}

	return results
}
