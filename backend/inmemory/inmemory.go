package inmemory

import (
	"context"
	g "github.com/Thooms/gatekeeper"
	"sync"
)

type Backend struct {
	data map[g.Key]g.Stats
	mu   *sync.Mutex
}

func New() *Backend {
	return &Backend{
		data: map[g.Key]g.Stats{},
		mu:   &sync.Mutex{},
	}
}

func (b *Backend) Allow(_ context.Context, k g.Key) (bool, g.Stats, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	s, ok := b.data[k]
	if !ok {
		return false, g.Stats{}, g.ErrUnknownKey
	}
	if s.Remaining > 0 {
		newS := g.Stats{Remaining: s.Remaining - 1, Limit: s.Limit}
		b.data[k] = newS
		return true, newS, nil
	}
	return false, s, nil
}

func (b *Backend) Stats(_ context.Context, k g.Key) (g.Stats, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	s, ok := b.data[k]
	if !ok {
		return g.Stats{}, g.ErrUnknownKey
	}
	return s, nil
}

// To be used for testing
func (b *Backend) Set(k g.Key, limit int64) {
	b.mu.Lock()
	b.data[k] = g.Stats{Remaining: limit, Limit: limit}
	b.mu.Unlock()
}
