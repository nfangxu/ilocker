package ilocker

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrHasLocked = errors.New("locked")
)

func newMeta(id string, releasedAt time.Time) *meta {
	return &meta{
		id:         id,
		releasedAt: releasedAt,
		mu:         &sync.RWMutex{},
	}
}

type meta struct {
	id         string
	releasedAt time.Time
	mu         *sync.RWMutex
}

func (m *meta) isExpired() bool {
	return m.releasedAt.Before(time.Now())
}

// Refresh the meta if the meta is expired
func (m *meta) Refresh(releasedAt time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	// if the meta is not expired, return false
	if !m.isExpired() {
		return false
	}
	// if the meta is expired, update the meta, return true
	m.releasedAt = releasedAt
	return true
}

// MemoryLocker is a simple in-memory implementation of the Locker interface.
type MemoryLocker struct {
	locks *sync.Map
}

func NewMemoryLocker(d time.Duration) ILocker {
	if d == 0 {
		d = time.Minute
	}
	locks := &sync.Map{}
	l := &MemoryLocker{locks: locks}
	go func(_l *MemoryLocker) {
		for {
			select {
			case <-time.After(d):
				locks.Range(func(key, value interface{}) bool {
					_m, ok := value.(*meta)
					if !ok {
						locks.Delete(key)
					}
					if _m.isExpired() {
						_ = _l.UnLock(context.Background(), _m.id)
					}
					return true
				})
			}
		}
	}(l)
	return l
}

func (m *MemoryLocker) Lock(ctx context.Context, id string, ttl time.Duration) (ILocked, error) {
	now := time.Now()
	value, ok := m.locks.LoadOrStore(id, newMeta(id, now.Add(ttl)))
	// If the key is not loaded, it is new, and we can lock it.
	if !ok {
		return Locked(m, id)
	}
	_m, ok := value.(*meta)
	// Retry if the value is not *meta
	if !ok {
		return m.Retry(ctx, id, ttl)
	}
	// Refresh the meta if the meta is expired
	if ok = _m.Refresh(now.Add(ttl)); !ok {
		return nil, ErrHasLocked
	}
	return Locked(m, id)
}

func (m *MemoryLocker) Locking(ctx context.Context, id string) bool {
	ld, ok := m.locks.Load(id)
	if !ok || ld == nil {
		return false
	}
	_m, ok := ld.(*meta)
	if !ok {
		return false
	}
	_m.mu.RLock()
	defer _m.mu.RUnlock()
	if _m.releasedAt.Before(time.Now()) {
		_ = m.UnLock(ctx, id)
		return false
	}
	return true
}

func (m *MemoryLocker) UnLock(ctx context.Context, id string) error {
	m.locks.Delete(id)
	return nil
}

func (m *MemoryLocker) Retry(ctx context.Context, id string, ttl time.Duration) (ILocked, error) {
	m.locks.Delete(id)
	return m.Lock(ctx, id, ttl)
}
