package package_manager

import (
	"TLExtractor/consts"
	"TLExtractor/utils"
	"TLExtractor/utils/package_manager/types"
	"fmt"
	"github.com/Laky-64/http"
	"io"
	"os"
	"path"
)

func download(info types.PackageInfo) error {
	sizeHuman := utils.HumanReadableBytes(int64(info.Size))
	fmt.Println(fmt.Sprintf("Downloading %s (%s)", info.FileName, sizeHuman))
	filePath := path.Join(
		consts.EnvFolder,
		consts.TempBins,
		info.FileName,
	)
	if _, err := os.Stat(filePath); err == nil {
		fmt.Println(
			fmt.Sprintf(
				"  Using cached %s (%s)",
				info.FileName,
				sizeHuman,
			),
		)
		return nil
	}
	pb := utils.NewProgressBar(info.Size)
	res, err := http.ExecuteRequest(
		info.DownloadURL,
		http.OverloadReader(
			func(r io.Reader) io.Reader {
				return pb.NewProxyReader(r)
			},
		),
	)
	pb.Finish()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(consts.EnvFolder, consts.TempBins), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	return os.WriteFile(
		filePath,
		res.Body,
		os.ModePerm,
	)
}
