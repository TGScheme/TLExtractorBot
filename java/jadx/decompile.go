package jadx

import (
	"TLExtractor/consts"
	"TLExtractor/utils/proxy_reader"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
)

func Decompile(callable func(percentage int64)) error {
	execPath, err := findExec()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(consts.EnvFolder, consts.TempDecompiled), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	java, err := javaExec()
	if err != nil {
		return err
	}
	cmd := exec.Command(
		java,
		"-cp",
		execPath,
		"jadx.cli.JadxCLI",
		"--comments-level",
		"none",
		"--no-replace-consts",
		"--no-res",
		"--no-inline-anonymous",
		"-j",
		strconv.Itoa(runtime.GOMAXPROCS(0)),
		"--output-dir",
		path.Join(consts.EnvFolder, consts.TempDecompiled),
		path.Join(consts.EnvFolder, consts.TempApk),
	)
	pb := proxy_reader.NewProxyReader(
		consts.UpdateMessageRate,
		100,
		func(percentage int64, _ int64) {
			callable(percentage)
		},
	)
	defer pb.Close()
	var terminal string
	compile := regexp.MustCompile(`INFO\s+-\s+progress:\s+[0-9]+\s+of\s+[0-9]+\s+\(([0-9]+)%\)`)
	pb.AttachCustomIncrementer(func(p []byte) int {
		terminal += string(p)
		res := compile.FindAllStringSubmatch(terminal, -1)
		if len(res) > 0 {
			percent, _ := strconv.Atoi(res[len(res)-1][1])
			return percent
		}
		return 0
	})
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = pb
	if err = cmd.Run(); err != nil {
		return errors.New(stdErr.String())
	}
	return nil
}
