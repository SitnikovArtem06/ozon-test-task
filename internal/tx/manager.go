package tx

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
type pgxTxManager struct {
	dbpool *pgxpool.Pool
}

func NewPgxTxManager(pool *pgxpool.Pool) TransactionManager {
	return &pgxTxManager{dbpool: pool}
}

type txKey struct{}
type withTxKey struct{}

func (p *pgxTxManager) Do(parent context.Context, withTx bool, fn func(ctx context.Context) error) error {

	if !withTx {
		ctx := context.WithValue(parent, withTxKey{}, false)

		err := fn(ctx)

		return err
	}

	tx, err := p.dbpool.Begin(parent)

	if err != nil {
		return err
	}

	ctx := context.WithValue(parent, withTxKey{}, true)

	ctx = context.WithValue(ctx, txKey{}, tx)

	err = fn(ctx)
	if err != nil {
		_ = tx.Rollback(parent)
		return err
	}

	if err = tx.Commit(parent); err != nil {
		return err
	}

	return nil

}

func (p *pgxTxManager) GetConnection(ctx context.Context) (DB, error) {

	withTx, _ := ctx.Value(withTxKey{}).(bool)

	if !withTx {
		return p.dbpool, nil
	}

	v := ctx.Value(txKey{})
	if v == nil {
		return nil, errors.New("tx not found in context")
	}

	tx, ok := v.(pgx.Tx)
	if !ok {
		return nil, errors.New("invalid tx in context")
	}

	return tx, nil

}
