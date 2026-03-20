package service

import "context"

type Repository interface {
	Create(ctx context.Context, shortUrl, url string) error
	GetByShort(ctx context.Context, shortUrl string) (string, error)
	GetByOrigin(ctx context.Context, url string) (string, error)
}
