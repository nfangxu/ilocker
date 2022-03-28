package ilocker

import (
	"context"
	"time"
)

type ILocker interface {
	Lock(ctx context.Context, id string, ttl time.Duration) (ILocked, error)
	Locking(ctx context.Context, id string) bool
	UnLock(ctx context.Context, id string) error
}

type ILocked interface {
	Locking(ctx context.Context) bool
	UnLock(ctx context.Context) error
}

// locked is a struct that implements the Locked interface.
type locked struct {
	locker ILocker
	id     string
}

func Locked(locker ILocker, id string) (ILocked, error) {
	return &locked{locker: locker, id: id}, nil
}

func (m *locked) Locking(ctx context.Context) bool {
	return m.locker.Locking(ctx, m.id)
}

func (m *locked) UnLock(ctx context.Context) error {
	return m.locker.UnLock(ctx, m.id)
}
