### Golang locker interface

- Use the built-in memory locker

```go
package main

import (
	"context"
	"github.com/nfangxu/ilocker"
	"time"
)

func main() {
	locker := ilocker.NewMemoryLocker(time.Minute)
	ld, err := locker.Lock(context.TODO(), "foo", time.Second)
	if err != nil {
		panic(err)
	}
	defer ld.UnLock(context.TODO())
	// TODO: do something
}
```

- Custom locker

```go
package main

import (
	"context"
	"github.com/nfangxu/ilocker"
	"time"
)

type RedisLocker struct {
}

func (r *RedisLocker) Lock(ctx context.Context, id string, ttl time.Duration) (ilocker.ILocked, error) {
	// TODO: implement
	return ilocker.Locked(r, id) // return a locked object
}

func (r *RedisLocker) Locking(ctx context.Context, id string) bool {
	// TODO: implement
	return false
}

func (r *RedisLocker) UnLock(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}

```