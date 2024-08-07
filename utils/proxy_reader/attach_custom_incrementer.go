package proxy_reader

func (b *ProxyReader) AttachCustomIncrementer(f func(data []byte) int) {
	b.customCallable = f
}
