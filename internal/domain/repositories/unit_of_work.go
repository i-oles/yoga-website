package repositories

import "context"

type IUnitOfWork interface {
	WithTransaction(ctx context.Context, fn func(r Repositories) error) error
}
