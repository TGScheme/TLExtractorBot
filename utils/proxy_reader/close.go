package proxy_reader

func (b *ProxyReader) Close() {
	if b != nil && b.callable != nil {
		b.processed = b.total
		b.callable(b.processed, b.total)
		b.finish <- true
	}
}
