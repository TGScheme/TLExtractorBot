package bot

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

type progressWriter struct {
	file     *os.File
	total    int64
	mutex    sync.Mutex
	written  int64
	reported int64
	lastAt   time.Time
	updates  chan<- int64
}

func (w *progressWriter) WriteAt(chunk []byte, offset int64) (int, error) {
	written, err := w.file.WriteAt(chunk, offset)
	if w.updates == nil || w.total <= 0 {
		return written, err
	}
	w.mutex.Lock()
	w.written += int64(written)
	percentage := w.written * 100 / w.total
	report := percentage > w.reported && time.Since(w.lastAt) >= consts.UpdateMessageRate
	if report {
		w.lastAt, w.reported = time.Now(), percentage
	}
	w.mutex.Unlock()
	if report {
		select {
		case w.updates <- percentage:
		default:
		}
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
	writer := &progressWriter{file: file, total: document.Size, lastAt: time.Now()}
	if onProgress != nil {
		updates := make(chan int64, 1)
		reported := make(chan struct{})
		go func() {
			defer close(reported)
			for percentage := range updates {
				onProgress(percentage)
			}
		}()
		writer.updates = updates
		defer func() {
			close(updates)
			<-reported
		}()
	}
	_, err = downloader.NewDownloader().Download(api, &tg.InputDocumentFileLocation{
		ID:            document.ID,
		AccessHash:    document.AccessHash,
		FileReference: document.FileReference,
	}).WithThreads(consts.DownloadThreads).WithRetryHandler(func(event downloader.RetryEvent) {
		gologging.Warn(fmt.Sprintf(
			"download: %s retried (attempt %d): %v", event.Operation, event.Attempt, event.Err,
		))
	}).Parallel(ctx.mtProtoCtx, writer)
	return err
}
