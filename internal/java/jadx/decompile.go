package jadx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
)

const maxReportedFailures = 10

var progressRgx = regexp.MustCompile(`INFO\s+-\s+progress:\s+[0-9]+\s+of\s+[0-9]+\s+\(([0-9]+)%\)`)

func Decompile(cfg *config.Config, onProgress func(percentage int64)) error {
	sources := path.Join(cfg.WorkDir, consts.TempSourcesRoot)
	if err := os.RemoveAll(path.Join(cfg.WorkDir, consts.TempDecompiled)); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.MkdirAll(sources, os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	threads := cfg.JadxThreads
	if threads < 1 {
		threads = runtime.NumCPU()
	}
	jvmOpts := strings.Fields(cfg.JadxJVMOpts)
	if !hasJVMOpt(jvmOpts, "-XX:ActiveProcessorCount") {
		jvmOpts = append(jvmOpts, "-XX:ActiveProcessorCount="+strconv.Itoa(threads))
	}
	inputs := selectInputs(path.Join(cfg.WorkDir, consts.TempApk), path.Join(cfg.WorkDir, consts.TempDex), consts.TgnetPackage)
	args := append(
		jvmOpts,
		"-Djdk.util.zip.disableZip64ExtraFieldValidation=true",
		"-cp", cfg.JadxJar+string(os.PathListSeparator)+cfg.ExtractJar,
		"TLExtract",
		strings.Join(inputs, string(os.PathListSeparator)),
		sources,
		consts.TgnetPackage,
		strconv.Itoa(threads),
	)
	gologging.Info(fmt.Sprintf(
		"jadx: %d threads (numcpu %d, gomaxprocs %d), jvm opts %q, package %s",
		threads, runtime.NumCPU(), runtime.GOMAXPROCS(0), strings.Join(jvmOpts, " "), consts.TgnetPackage,
	))
	started := time.Now()
	cmd := exec.Command(cfg.JavaBin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	if err = cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(scanLines)
	var last int64 = -1
	var lastAt time.Time
	var failures []string
	for scanner.Scan() {
		line := scanner.Text()
		match := progressRgx.FindStringSubmatch(line)
		if match == nil {
			if strings.HasPrefix(line, "ERROR") {
				failures = append(failures, line)
				if len(failures) > maxReportedFailures {
					failures = failures[1:]
				}
			}
			continue
		}
		percentage, _ := strconv.ParseInt(match[1], 10, 64)
		if percentage == last || time.Since(lastAt) < consts.UpdateMessageRate {
			continue
		}
		last, lastAt = percentage, time.Now()
		onProgress(percentage)
	}
	if err = scanner.Err(); err != nil {
		gologging.Warn("jadx: stopped reading the output:", err)
		_, _ = io.Copy(io.Discard, stdout)
	}

	if err = cmd.Wait(); err != nil {
		if message := stdErr.String(); len(message) > 0 {
			return errors.New(message)
		}
		if len(failures) > 0 {
			return fmt.Errorf("%w: %s", err, strings.Join(failures, "\n"))
		}
		return err
	}
	gologging.Info(fmt.Sprintf("jadx: decompiled the package in %s", time.Since(started).Round(time.Millisecond)))
	return checkOutput(cfg.WorkDir)
}

func hasJVMOpt(opts []string, name string) bool {
	for _, opt := range opts {
		if opt == name || strings.HasPrefix(opt, name+"=") || strings.HasPrefix(opt, name+":") {
			return true
		}
	}
	return false
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
