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

func (r *Registry) Check(ctx context.Context) map[string]string {
	results := make(map[string]string)

	for _, c := range r.checkers {

		if err := c.Check(ctx); err != nil {
			results[c.Name()] = err.Error()
			continue
		}

		results[c.Name()] = "ok"
	}

	return results
}
