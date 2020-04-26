package commands

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Upload struct {
	url         string
	bodyReader  io.Reader
	contentType string
}

func NewUpload(u string, br io.Reader, ct string) *Upload {
	return &Upload{
		url:         u,
		bodyReader:  br,
		contentType: ct,
	}
}

func (u *Upload) Do() error {
	request, err := http.NewRequest(http.MethodPost, u.url, u.bodyReader)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", u.contentType)
	request.Close = true

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to upload file: %s", errorMessage(response.Body))
	}

	return nil
}

func MakeMultipartReader(files []string) (io.Reader, string, int64, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	var size int64

	for _, fileName := range files {
		fi, err := os.Stat(fileName)
		if err != nil {
			return nil, "", 0, err
		}
		if fi.IsDir() {
			continue
		}

		size += fi.Size()

		file, err := os.Open(fileName)
		if err != nil {
			return nil, "", 0, err
		}

		baseName := filepath.Base(fileName)
		part, err := writer.CreateFormFile(baseName, baseName)
		if err != nil {
			return nil, "", 0, err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, "", 0, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", 0, err
	}

	return body, writer.FormDataContentType(), size, nil
}
