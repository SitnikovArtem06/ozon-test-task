package tx

import (
	"context"
	"sync"
)

type inMemoryTxManager struct {
	mu sync.Mutex
}

func NewInMemoryTxManager() TransactionManager {
	return &inMemoryTxManager{}
}

func (m *inMemoryTxManager) Do(parent context.Context, withTx bool, fn func(ctx context.Context) error) error {

	if withTx {
		m.mu.Lock()
		defer m.mu.Unlock()
	}

	ctx := context.WithValue(parent, withTxKey{}, false)
	return fn(ctx)
}

// не используется(заглушка)
func (m *inMemoryTxManager) GetConnection(ctx context.Context) (DB, error) {
	return nil, nil
}
