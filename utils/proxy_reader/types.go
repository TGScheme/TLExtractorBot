package proxy_reader

import (
	"io"
	"time"
)

type ProgressCallable func(downloadedBytes int64, totalBytes int64)

type ProxyReader struct {
	io.Reader
	io.Writer
	processed      int64
	lastProcessed  int64
	callable       ProgressCallable
	finish         chan bool
	refreshRate    time.Duration
	total          int64
	customCallable func(data []byte) int
}
