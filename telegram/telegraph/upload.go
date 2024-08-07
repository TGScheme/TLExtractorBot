package telegraph

import (
	"TLExtractor/consts"
	"TLExtractor/http"
	multipart "TLExtractor/http/types"
	"encoding/json"
	"fmt"
)

func upload(content []byte, fileType string) (string, error) {
	res := http.ExecuteRequest(
		fmt.Sprintf("%s/upload", consts.TelegraphUrl),
		http.Method("POST"),
		http.Headers(map[string]string{"Content-Type": fileType}),
		http.MultiPartForm(
			http.Files(map[string]multipart.FileDescriptor{
				"file": {
					FileName: "blob",
					Content:  content,
				},
			}),
		),
	)
	if res.Error != nil {
		return "", res.Error
	}
	var parsedContent []map[string]string
	err := json.Unmarshal(res.Read(), &parsedContent)
	if err != nil {
		return "", err
	}
	return consts.TelegraphUrl + parsedContent[0]["src"], nil
}
