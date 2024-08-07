package types

import (
	io2 "TLExtractor/io"
	"io"
)

type HTTPResult struct {
	Body       io.Reader
	Error      error
	StatusCode int
	cacheRead  []byte
}

func (r *HTTPResult) SetFallback(body []byte) {
	r.cacheRead = body
	r.Body = nil
}

func (r *HTTPResult) Read() []byte {
	if r.cacheRead != nil {
		return r.cacheRead
	}
	if r.Body == nil {
		return nil
	}
	buf, err := io2.ReadFile(r.Body)
	if err != nil {
		return nil
	}
	defer func() {
		r.Body = nil
	}()
	r.cacheRead = buf
	return r.cacheRead
}

func (r *HTTPResult) ReadString() string {
	return string(r.Read())
}
