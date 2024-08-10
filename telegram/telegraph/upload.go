package telegraph

import (
	"TLExtractor/consts"
	"encoding/json"
	"fmt"
	"github.com/Laky-64/http"
	multipart "github.com/Laky-64/http/types"
)

func upload(content []byte, fileType string) (string, error) {
	res, err := http.ExecuteRequest(
		fmt.Sprintf("%s/upload", consts.TelegraphUrl),
		http.Method("POST"),
		http.Headers(map[string]string{"Content-Type": fileType}),
		http.MultiPartForm(
			nil,
			map[string]multipart.FileDescriptor{
				"file": {
					FileName: "blob",
					Content:  content,
				},
			},
		),
	)
	if err != nil {
		return "", err
	}
	var parsedContent []map[string]string
	err = json.Unmarshal(res.Body, &parsedContent)
	if err != nil {
		return "", err
	}
	return consts.TelegraphUrl + parsedContent[0]["src"], nil
}
