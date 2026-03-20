package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const (
	TimeOut       = 5
	RetryAttempts = 5
	RetryDelay    = time.Second
)

func InitDb(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	var lastErr error

	for i := 1; i <= RetryAttempts; i++ {

		tctx, cancel := context.WithTimeout(ctx, TimeOut*time.Second)

		dbpool, err := pgxpool.New(tctx, connStr)
		if err != nil {
			lastErr = fmt.Errorf("new pool fail: %w", err)

		} else {
			if err = dbpool.Ping(tctx); err != nil {
				dbpool.Close()
				lastErr = fmt.Errorf("db ping: %w", err)
			} else {
				cancel()
				return dbpool, nil
			}
		}

		cancel()

		if i < RetryAttempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(RetryDelay):

			}
		}

	}

	return nil, fmt.Errorf("db connect failed after %d attempts: %w", RetryAttempts, lastErr)
}
