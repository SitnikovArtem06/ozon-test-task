package tx

import "context"

type TransactionManager interface {
	Do(parent context.Context, withTx bool, fn func(ctx context.Context) error) error
	GetConnection(ctx context.Context) (DB, error)
}
