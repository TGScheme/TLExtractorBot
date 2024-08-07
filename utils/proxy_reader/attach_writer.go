package proxy_reader

import "io"

func (b *ProxyReader) AttachWriter(writer io.Writer) *ProxyReader {
	b.Writer = writer
	return b
}
