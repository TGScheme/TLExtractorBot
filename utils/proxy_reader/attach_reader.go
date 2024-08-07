package proxy_reader

import "io"

func (b *ProxyReader) AttachReader(reader io.Reader) *ProxyReader {
	b.Reader = reader
	return b
}
