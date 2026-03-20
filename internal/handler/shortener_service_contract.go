package handler

import "context"

type shortenerService interface {
	AddOrigin(ctx context.Context, originUrl string) (string, error)
	GetOrigin(ctx context.Context, shortUrl string) (string, error)
}
