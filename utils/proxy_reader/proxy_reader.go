package proxy_reader

import (
	"time"
)

func NewProxyReader(refreshRate time.Duration, total int64, callable ProgressCallable) *ProxyReader {
	b := &ProxyReader{
		callable:    callable,
		finish:      make(chan bool),
		refreshRate: refreshRate,
		total:       total,
	}
	if callable != nil {
		go func() {
			for {
				select {
				case <-b.finish:
					return
				case <-time.After(b.refreshRate):
					if b.processed < b.total && b.processed != b.lastProcessed {
						b.lastProcessed = b.processed
						b.callable(b.processed, b.total)
					}
				}
			}
		}()
	}
	return b
}
