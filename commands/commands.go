package commands

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func makeBodyReader(files []string) (io.Reader, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, fileName := range files {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, "", err
		}

		baseName := filepath.Base(fileName)
		part, err := writer.CreateFormFile(baseName, baseName)
		if err != nil {
			return nil, "", err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return nil, "", err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func Upload(url string, files []string) error {
	fmt.Println(url)
	body, contentType, err := makeBodyReader(files)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", contentType)
	request.Close = true

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to upload file: %s", response.Status)
	}

	return nil
}

func makeFileName(url string) (string, error) {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	fileName := url[strings.LastIndex(url, "/")+1:]

	counter := 0
	for {
		_, err := os.Open(fileName)
		if os.IsNotExist(err) {
			break
		}

		if err != nil {
			return "", err
		}

		fileName = fmt.Sprintf("%s_%d", fileName, counter)
		counter++
	}

	return fileName, nil
}

func Download(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unable to download file: %s", response.Status)
	}

	fileName, err := makeFileName(response.Request.URL.String())
	if err != nil {
		return "", err
	}

	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(f, response.Body)
	if err != nil {
		return "", err
	}

	return fileName, nil
}
