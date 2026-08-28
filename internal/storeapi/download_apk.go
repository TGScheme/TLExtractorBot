package storeapi

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/storeapi/types"
)

type countingReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	lastAt     time.Time
	onProgress func(percentage int64)
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	c.downloaded += int64(n)
	if c.total > 0 && c.onProgress != nil && time.Since(c.lastAt) >= consts.UpdateMessageRate {
		c.lastAt = time.Now()
		c.onProgress(c.downloaded * 100 / c.total)
	}
	return n, err
}

func apkSize(uri string) int64 {
	res, err := http.ExecuteRequest(uri, http.Method("HEAD"))
	if err != nil {
		return 0
	}
	values, ok := res.Headers["Content-Length"]
	if !ok || len(values) == 0 {
		return 0
	}
	size, err := strconv.ParseInt(values[0], 10, 64)
	if err != nil {
		return 0
	}
	return size
}

func DownloadApk(cfg *config.Config, info *types.AppInfo, onProgress func(percentage int64)) error {
	uri := fmt.Sprintf("%s&version=%d", info.FileURL, time.Now().Second())
	counter := &countingReader{total: apkSize(uri), onProgress: onProgress}
	res, err := http.ExecuteRequest(uri, http.OverloadReader(func(r io.Reader) io.Reader {
		counter.reader = r
		return counter
	}))
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(cfg.WorkDir, consts.TempBins), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	return os.WriteFile(path.Join(cfg.WorkDir, consts.TempApk), res.Body, os.ModePerm)
}
