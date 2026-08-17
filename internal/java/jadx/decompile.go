package jadx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
)

var progressRgx = regexp.MustCompile(`INFO\s+-\s+progress:\s+[0-9]+\s+of\s+[0-9]+\s+\(([0-9]+)%\)`)

func Decompile(cfg *config.Config, onProgress func(percentage int64)) error {
	outDir := path.Join(cfg.WorkDir, consts.TempDecompiled)
	if err := os.RemoveAll(outDir); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	cmd := exec.Command(
		cfg.JavaBin,
		"-Xms256M",
		"-XX:MaxRAMPercentage=70.0",
		"-Djdk.util.zip.disableZip64ExtraFieldValidation=true",
		"-cp", cfg.JadxJar,
		"jadx.cli.JadxCLI",
		"--comments-level", "none",
		"--no-replace-consts",
		"--no-res",
		"--no-inline-anonymous",
		"-j", strconv.Itoa(runtime.GOMAXPROCS(0)),
		"--output-dir", outDir,
		path.Join(cfg.WorkDir, consts.TempApk),
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	if err = cmd.Start(); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(stdout)
		scanner.Split(scanLines)
		var last int64 = -1
		var lastAt time.Time
		for scanner.Scan() {
			match := progressRgx.FindStringSubmatch(scanner.Text())
			if match == nil {
				continue
			}
			percentage, _ := strconv.ParseInt(match[1], 10, 64)
			if percentage == last || time.Since(lastAt) < consts.UpdateMessageRate {
				continue
			}
			last, lastAt = percentage, time.Now()
			onProgress(percentage)
		}
	}()
	<-done

	if err = cmd.Wait(); err != nil {
		if message := stdErr.String(); len(message) > 0 {
			return errors.New(message)
		}
		return err
	}
	return checkOutput(cfg.WorkDir)
}

func checkOutput(workDir string) error {
	sources := path.Join(workDir, consts.TempSources)
	entries, err := os.ReadDir(sources)
	if err != nil {
		return fmt.Errorf("jadx produced no %s: %w", consts.TempSources, err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "TLRPC") {
			return nil
		}
	}
	return fmt.Errorf("jadx produced no TLRPC class in %s (%d entries)", sources, len(entries))
}

func scanLines(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
