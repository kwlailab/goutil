package shutdown

import (
	"context"
	"sync"
	"time"

	"github.com/kwlailab/goutil/pkg/log"
)

type Fn func(context.Context) error

var fns []Fn

func Register(fn Fn) {
	fns = append(fns, fn)
}

func Graceful(timeout time.Duration) {
	cctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.InfoCtx(cctx, "[shutdown] graceful stopping...")
	wg := &sync.WaitGroup{}

	for i := len(fns) - 1; i >= 0; i-- {
		fn := fns[i]
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(fn Fn) {
			defer wg.Done()
			if err := fn(cctx); err != nil {
				log.ErrorCtx(cctx, "[shutdown] shutdown failed", "err", err)
			}
		}(fn)
	}
	wg.Wait()
	log.InfoCtx(cctx, "[shutdown] server graceful stopped")
}
