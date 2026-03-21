package repository

import (
	"URLshortener/internal/tx"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository struct {
	tm tx.TransactionManager
}

func NewPostgresRepository(tm tx.TransactionManager) *PostgresRepository {
	return &PostgresRepository{tm: tm}
}

const PG_UNIQUE_VIOLATION = "23505"

func (r *PostgresRepository) Create(ctx context.Context, shortUrl, url string) error {

	conn, err := r.tm.GetConnection(ctx)
	if err != nil {
		return err
	}
	if _, err = conn.Exec(ctx, `INSERT INTO links (short_url, original_url) VALUES ($1, $2)`, shortUrl, url); err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == PG_UNIQUE_VIOLATION {
			switch pgErr.ConstraintName {
			case "links_pkey":
				return ErrDuplicateShort
			case "links_original_url_key":
				return ErrDuplicateOrigin
			}
		}
		return fmt.Errorf("create link: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByShort(ctx context.Context, shortUrl string) (string, error) {
	var originalURL string

	conn, err := r.tm.GetConnection(ctx)
	if err != nil {
		return "", err
	}

	if err = conn.QueryRow(ctx, `SELECT original_url FROM links WHERE short_url = $1`, shortUrl).Scan(&originalURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFoundShort
		}
		return "", fmt.Errorf("get by short: %w", err)
	}

	return originalURL, nil
}

func (r *PostgresRepository) GetByOrigin(ctx context.Context, url string) (string, error) {
	var shortURL string

	conn, err := r.tm.GetConnection(ctx)
	if err != nil {
		return "", err
	}

	if err = conn.QueryRow(ctx, `SELECT short_url FROM links WHERE original_url = $1`, url).Scan(&shortURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFoundOrigin
		}
		return "", fmt.Errorf("get by origin: %w", err)
	}

	return shortURL, nil
}
