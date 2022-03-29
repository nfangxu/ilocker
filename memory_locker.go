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

type meta struct {
	id         string
	releasedAt time.Time
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
	go func() {
		for {
			select {
			case <-time.After(d):
				locks.Range(func(key, value interface{}) bool {
					if value.(*meta).releasedAt.Before(time.Now()) {
						locks.Delete(key)
					}
					return true
				})
			}
		}
	}()
	return &MemoryLocker{locks: locks}
}

func (m *MemoryLocker) Lock(ctx context.Context, id string, ttl time.Duration) (ILocked, error) {
	now := time.Now()
	_m, ok := m.locks.LoadOrStore(id, &meta{id: id, releasedAt: now.Add(ttl)})
	// If the key is not loaded, it is new, and we can lock it.
	if !ok {
		return Locked(m, id)
	}
	// If the key is loaded, and it is not expired, we cannot lock it.
	if _m.(*meta).releasedAt.After(now) {
		return nil, ErrHasLocked
	}
	// If the key is loaded, and it is expired, we can lock it. And we need to update the releasedAt.
	_m.(*meta).releasedAt = now.Add(ttl)
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
