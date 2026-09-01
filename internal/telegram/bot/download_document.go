package bot

import (
	"os"
	"sync"
	"time"

	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

type progressWriter struct {
	file       *os.File
	total      int64
	mutex      sync.Mutex
	written    int64
	lastAt     time.Time
	onProgress func(percentage int64)
}

func (w *progressWriter) WriteAt(chunk []byte, offset int64) (int, error) {
	written, err := w.file.WriteAt(chunk, offset)
	if w.onProgress == nil || w.total <= 0 {
		return written, err
	}
	w.mutex.Lock()
	w.written += int64(written)
	report := time.Since(w.lastAt) >= consts.UpdateMessageRate
	percentage := w.written * 100 / w.total
	if report {
		w.lastAt = time.Now()
	}
	w.mutex.Unlock()
	if report {
		w.onProgress(percentage)
	}
	return written, err
}

func (ctx *Client) DownloadDocument(document *tg.Document, dest string, onProgress func(percentage int64)) error {
	api, err := ctx.mtProtoClient()
	if err != nil {
		return err
	}
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	writer := &progressWriter{
		file:       file,
		total:      document.Size,
		lastAt:     time.Now(),
		onProgress: onProgress,
	}
	_, err = downloader.NewDownloader().Download(api, &tg.InputDocumentFileLocation{
		ID:            document.ID,
		AccessHash:    document.AccessHash,
		FileReference: document.FileReference,
	}).Parallel(ctx.mtProtoCtx, writer)
	return err
}
