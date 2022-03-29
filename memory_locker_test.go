package ilocker

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestMemoryLocker_Lock(t *testing.T) {
	l := NewMemoryLocker(0)
	ctx := context.Background()
	ld1, err := l.Lock(ctx, "foo", time.Second)
	assert.NoErrorf(t, err, "Lock() should not return error")
	assert.True(t, ld1.Locking(ctx), "Locking() should return true")
	time.Sleep(time.Second)
	assert.False(t, ld1.Locking(ctx), "Locking() should return false")

	ld2, err := l.Lock(ctx, "bar", time.Second)
	assert.NoErrorf(t, err, "Lock() should not return error")
	assert.True(t, ld2.Locking(ctx), "Locking() should return true")
	err = ld2.UnLock(ctx)
	assert.NoErrorf(t, err, "UnLock() should not return error")
	assert.False(t, ld2.Locking(ctx), "Locking() should return false")
}

func TestMemoryLocker_Lock_Async(t *testing.T) {
	l := NewMemoryLocker(0)
	ctx := context.Background()

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func(_wg *sync.WaitGroup) {
		ld1, err := l.Lock(ctx, "foo", time.Second)
		assert.NoErrorf(t, err, "Lock() should not return error")
		assert.True(t, ld1.Locking(ctx), "Locking() should return true")
		time.Sleep(time.Second)
		assert.False(t, ld1.Locking(ctx), "Locking() should return false")
		_wg.Done()
	}(&wg)
	go func(_wg *sync.WaitGroup) {
		time.Sleep(time.Second / 2)
		_, err := l.Lock(ctx, "foo", time.Second)
		assert.EqualError(t, err, ErrHasLocked.Error(), "Lock() should return ErrHasLocked")
		time.Sleep(time.Second)
		ld2, err := l.Lock(ctx, "foo", time.Second)
		assert.True(t, ld2.Locking(ctx), "Locking() should return true")
		err = ld2.UnLock(ctx)
		assert.NoErrorf(t, err, "UnLock() should not return error")
		assert.False(t, ld2.Locking(ctx), "Locking() should return false")
		_wg.Done()
	}(&wg)
	wg.Wait()
}

func TestMemoryLocker_Clean(t *testing.T) {
	l := NewMemoryLocker(time.Second)
	ctx := context.Background()
	_, err := l.Lock(ctx, "clean", time.Second)
	assert.NoErrorf(t, err, "Lock() should not return error")

	_l := l.(*MemoryLocker)
	_, ok := _l.locks.Load("clean")
	assert.True(t, ok, "locks.Load() should return true")
	time.Sleep(time.Second * 2)
	_, ok = _l.locks.Load("clean")
	assert.False(t, ok, "locks.Load() should return false")

	// Clean another one
	l.Lock(ctx, "clean", time.Second)
	time.Sleep(time.Second * 2)
	_, ok = _l.locks.Load("clean")
	assert.False(t, ok, "locks.Load() should return false")
}
