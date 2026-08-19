package postgres

import (
	"context"
)

func (r Repository) withQueryTimeout(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		ctx,
		r.queryTimeout,
	)
}
