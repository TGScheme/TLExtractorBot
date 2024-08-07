package proxy_reader

func (b *ProxyReader) Write(p []byte) (n int, err error) {
	if b.Writer != nil {
		n, err = b.Writer.Write(p)
	} else {
		n = len(p)
	}
	if b.customCallable != nil {
		b.processed = int64(b.customCallable(p))
	} else {
		b.processed += int64(n)
	}
	return
}
