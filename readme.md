### Golang locker interface

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
