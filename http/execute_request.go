package http

import (
	"TLExtractor/http/types"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func ExecuteRequest(url string, options ...RequestOption) *types.HTTPResult {
	var opt types.RequestOptions
	bodyRes := &types.HTTPResult{}
	for _, option := range options {
		option.Apply(&opt)
	}
	if (opt.MultiPart != nil) == (opt.Body != nil) && opt.MultiPart != nil {
		bodyRes.Error = fmt.Errorf("can't use multipart and body at the same time")
		return bodyRes
	}
	if opt.Method == "" {
		opt.Method = "GET"
	}
	client := http.Client{}
	var body io.Reader
	var multiPartWriter *multipart.Writer
	if opt.MultiPart != nil {
		reader := &bytes.Buffer{}
		multiPartWriter = multipart.NewWriter(reader)
		for k, v := range opt.MultiPart.Files {
			file, err := multiPartWriter.CreateFormFile(k, v.FileName)
			if err != nil {
				bodyRes.Error = err
				return bodyRes
			}
			_, err = file.Write(v.Content)
			if err != nil {
				bodyRes.Error = err
				return bodyRes
			}
		}
		_ = multiPartWriter.Close()
		body = reader
	} else if opt.Body != nil {
		body = bytes.NewBuffer(opt.Body)
	}
	req, err := http.NewRequest(opt.Method, url, body)
	if err != nil {
		bodyRes.Error = err
		return bodyRes
	}
	if opt.Headers != nil {
		for k, v := range opt.Headers {
			req.Header.Set(k, v)
		}
	}
	if multiPartWriter != nil {
		req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())
	}
	req.Header.Add("Accept-Encoding", "identity")
	do, err := client.Do(req)
	if err != nil {
		bodyRes.Error = err
		return bodyRes
	}
	var proxyBody io.Reader = do.Body
	if opt.OverloadReader != nil {
		proxyBody = opt.OverloadReader(proxyBody)
	}
	bodyRes.StatusCode = do.StatusCode
	if do.StatusCode == http.StatusForbidden || do.StatusCode == http.StatusNotFound {
		bodyRes.Error = fmt.Errorf("status code %d", do.StatusCode)
		return bodyRes
	}
	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusCreated && do.StatusCode != http.StatusNoContent {
		bodyRes.Error = fmt.Errorf("http status code %d", do.StatusCode)
		return bodyRes
	}
	bodyRes.Body = proxyBody
	bodyRes.Read()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(do.Body)
	return bodyRes
}
