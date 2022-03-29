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
	if m.Locking(ctx, id) {
		return nil, ErrHasLocked
	}
	m.locks.Store(id, &meta{id: id, releasedAt: time.Now().Add(ttl)})
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
