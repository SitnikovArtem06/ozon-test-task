package in_memory

import (
	"URLshortener/internal/repository"
	"context"
	"sync"
)

type InMemoryRepository struct {
	mu  sync.Mutex
	lru *lruCache
}

func NewInMemoryRepository(capacity int) *InMemoryRepository {
	return &InMemoryRepository{lru: newLRUCache(capacity)}
}

func (r *InMemoryRepository) Create(ctx context.Context, shortUrl, url string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lru.Put(shortUrl, url)
	return nil
}
func (r *InMemoryRepository) GetByShort(ctx context.Context, shortUrl string) (string, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if origin, ok := r.lru.GetByShort(shortUrl); !ok {
		return "", repository.ErrNotFoundShort
	} else {
		return origin, nil
	}

}

func (r *InMemoryRepository) GetByOrigin(ctx context.Context, url string) (string, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if origin, ok := r.lru.GetByOrigin(url); !ok {
		return "", repository.ErrNotFoundOrigin
	} else {
		return origin, nil
	}

}
