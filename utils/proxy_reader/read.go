package proxy_reader

func (b *ProxyReader) Read(p []byte) (n int, err error) {
	if b.Reader != nil {
		n, err = b.Reader.Read(p)
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
